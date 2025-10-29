/*
Copyright 2025 roguepikachu.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kyklosv1alpha1 "github.com/roguepikachu/kyklos/api/v1alpha1"
	"github.com/roguepikachu/kyklos/internal/engine"
)

// TimeWindowScalerReconciler reconciles a TimeWindowScaler object
type TimeWindowScalerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Clock    engine.Clock
}

// +kubebuilder:rbac:groups=kyklos.kyklos.io,resources=timewindowscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kyklos.kyklos.io,resources=timewindowscalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kyklos.kyklos.io,resources=timewindowscalers/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments/scale,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *TimeWindowScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the TimeWindowScaler instance
	tws := &kyklosv1alpha1.TimeWindowScaler{}
	if err := r.Get(ctx, req.NamespacedName, tws); err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, could have been deleted
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Initialize clock if not set (for testing)
	if r.Clock == nil {
		r.Clock = engine.RealClock{}
	}

	// Determine target namespace
	targetNamespace := tws.Spec.TargetRef.Namespace
	if targetNamespace == "" {
		targetNamespace = tws.Namespace
	}

	// Fetch the target Deployment
	deployment := &appsv1.Deployment{}
	deploymentKey := types.NamespacedName{
		Name:      tws.Spec.TargetRef.Name,
		Namespace: targetNamespace,
	}

	if err := r.Get(ctx, deploymentKey, deployment); err != nil {
		if apierrors.IsNotFound(err) {
			// Target not found - update status and requeue
			return r.handleMissingTarget(ctx, tws)
		}
		return ctrl.Result{}, err
	}

	// Check if paused - compute but don't apply
	if tws.Spec.Pause {
		logger.Info("TimeWindowScaler is paused",
			"name", tws.Name,
			"namespace", tws.Namespace)
		// Still compute to show what would happen
		return r.computeAndUpdateStatus(ctx, tws, deployment)
	}

	// Compute effective replicas using the engine
	engineInput := r.buildEngineInput(tws)
	engineOutput, err := engine.ComputeEffectiveReplicas(engineInput)
	if err != nil {
		logger.Error(err, "Failed to compute effective replicas")
		return ctrl.Result{}, err
	}

	// Log the decision
	logger.Info("Computed scaling decision",
		"nowLocal", engineInput.Now.In(mustLoadLocation(tws.Spec.Timezone)).Format(time.RFC3339),
		"nextBoundary", engineOutput.NextBoundary.Format(time.RFC3339),
		"effectiveReplicas", engineOutput.EffectiveReplicas,
		"currentWindow", engineOutput.CurrentWindow,
		"reason", engineOutput.Reason)

	// Compare with current state
	currentReplicas := *deployment.Spec.Replicas
	targetReplicas := engineOutput.EffectiveReplicas

	// Scale if needed
	if currentReplicas != targetReplicas && !tws.Spec.Pause {
		if err := r.scaleDeployment(ctx, deployment, targetReplicas); err != nil {
			return ctrl.Result{}, err
		}

		// Emit event
		eventType := "ScaledUp"
		if targetReplicas < currentReplicas {
			eventType = "ScaledDown"
		}
		r.Recorder.Event(tws, corev1.EventTypeNormal, eventType,
			fmt.Sprintf("Scaled from %d to %d replicas (window: %s)",
				currentReplicas, targetReplicas, engineOutput.CurrentWindow))
	}

	// Update status
	tws.Status.ObservedGeneration = tws.Generation
	tws.Status.EffectiveReplicas = &engineOutput.EffectiveReplicas
	tws.Status.TargetObservedReplicas = deployment.Spec.Replicas
	tws.Status.CurrentWindow = engineOutput.CurrentWindow
	nextBoundaryTime := metav1.NewTime(engineOutput.NextBoundary)
	tws.Status.NextBoundary = &nextBoundaryTime

	if currentReplicas == targetReplicas || tws.Spec.Pause {
		tws.Status.LastScaleTime = &metav1.Time{Time: r.Clock.Now()}
	}

	// Set Ready condition
	readyCondition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		ObservedGeneration: tws.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             "Reconciled",
		Message:            fmt.Sprintf("TimeWindowScaler is ready, window: %s", engineOutput.CurrentWindow),
	}

	meta.SetStatusCondition(&tws.Status.Conditions, readyCondition)

	// Update the status
	if err := r.Status().Update(ctx, tws); err != nil {
		return ctrl.Result{}, err
	}

	// Calculate requeue time - requeue just before next boundary
	requeueAfter := engineOutput.NextBoundary.Sub(r.Clock.Now()) - 10*time.Second
	if requeueAfter < 30*time.Second {
		requeueAfter = 30 * time.Second
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

// buildEngineInput converts TWS spec to engine input
func (r *TimeWindowScalerReconciler) buildEngineInput(tws *kyklosv1alpha1.TimeWindowScaler) engine.Input {
	windows := make([]engine.WindowSpec, len(tws.Spec.Windows))
	for i, w := range tws.Spec.Windows {
		windows[i] = engine.WindowSpec{
			Start:    w.Start,
			End:      w.End,
			Replicas: w.Replicas,
			Name:     w.Name,
			Days:     w.Days,
		}
	}

	input := engine.Input{
		Now:             r.Clock.Now(),
		Timezone:        tws.Spec.Timezone,
		Windows:         windows,
		DefaultReplicas: tws.Spec.DefaultReplicas,
		HolidayMode:     tws.Spec.HolidayMode,
		IsHoliday:       false, // TODO: check holiday ConfigMap
		Pause:           tws.Spec.Pause,
	}

	if tws.Spec.GracePeriodSeconds != nil {
		input.GracePeriodSecs = *tws.Spec.GracePeriodSeconds
	}

	if tws.Status.LastScaleTime != nil {
		input.LastScaleTime = &tws.Status.LastScaleTime.Time
	}

	if tws.Status.TargetObservedReplicas != nil {
		input.CurrentReplicas = *tws.Status.TargetObservedReplicas
	}

	return input
}

// scaleDeployment patches the deployment with new replica count
func (r *TimeWindowScalerReconciler) scaleDeployment(ctx context.Context, deployment *appsv1.Deployment, replicas int32) error {
	deployment.Spec.Replicas = &replicas
	return r.Update(ctx, deployment)
}

// handleMissingTarget handles case when target deployment is not found
func (r *TimeWindowScalerReconciler) handleMissingTarget(ctx context.Context, tws *kyklosv1alpha1.TimeWindowScaler) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Target deployment not found",
		"target", tws.Spec.TargetRef.Name,
		"namespace", tws.Spec.TargetRef.Namespace)

	// Set Degraded condition
	degradedCondition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		ObservedGeneration: tws.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             "TargetNotFound",
		Message:            fmt.Sprintf("Target deployment %s not found", tws.Spec.TargetRef.Name),
	}

	meta.SetStatusCondition(&tws.Status.Conditions, degradedCondition)
	tws.Status.ObservedGeneration = tws.Generation

	if err := r.Status().Update(ctx, tws); err != nil {
		return ctrl.Result{}, err
	}

	// Requeue after 5 minutes to check if target appears
	return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}

// computeAndUpdateStatus computes status when paused
func (r *TimeWindowScalerReconciler) computeAndUpdateStatus(ctx context.Context, tws *kyklosv1alpha1.TimeWindowScaler, deployment *appsv1.Deployment) (ctrl.Result, error) {
	engineInput := r.buildEngineInput(tws)
	engineOutput, err := engine.ComputeEffectiveReplicas(engineInput)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Update status but don't scale
	tws.Status.ObservedGeneration = tws.Generation
	tws.Status.EffectiveReplicas = &engineOutput.EffectiveReplicas
	tws.Status.TargetObservedReplicas = deployment.Spec.Replicas
	tws.Status.CurrentWindow = engineOutput.CurrentWindow
	nextBoundaryTime := metav1.NewTime(engineOutput.NextBoundary)
	tws.Status.NextBoundary = &nextBoundaryTime

	// Set Ready condition with paused state
	readyCondition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		ObservedGeneration: tws.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             "Paused",
		Message:            "TimeWindowScaler is paused",
	}

	meta.SetStatusCondition(&tws.Status.Conditions, readyCondition)

	if err := r.Status().Update(ctx, tws); err != nil {
		return ctrl.Result{}, err
	}

	// Still requeue to update status
	requeueAfter := engineOutput.NextBoundary.Sub(r.Clock.Now()) - 10*time.Second
	if requeueAfter < 30*time.Second {
		requeueAfter = 30 * time.Second
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

// mustLoadLocation loads a timezone location, panics on error (should not happen with validated input)
func mustLoadLocation(tz string) *time.Location {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		// This should not happen as timezone is validated by CRD
		return time.UTC
	}
	return loc
}

// SetupWithManager sets up the controller with the Manager.
func (r *TimeWindowScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kyklosv1alpha1.TimeWindowScaler{}).
		Named("timewindowscaler").
		Complete(r)
}
