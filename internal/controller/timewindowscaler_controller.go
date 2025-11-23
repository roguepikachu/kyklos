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
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kyklosv1alpha1 "github.com/roguepikachu/kyklos/api/v1alpha1"
	"github.com/roguepikachu/kyklos/internal/engine"
	"github.com/roguepikachu/kyklos/internal/metrics"
)

// Finalizer for cleaning up resources
const timeWindowScalerFinalizer = "kyklos.kyklos.io/finalizer"

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
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *TimeWindowScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	logger := log.FromContext(ctx)

	// Track reconciliation duration
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.ReconcileDurationSeconds.WithLabelValues(
			req.Namespace,
			req.Name,
			status,
		).Observe(duration)
	}()

	// Fetch the TimeWindowScaler instance
	tws := &kyklosv1alpha1.TimeWindowScaler{}
	if err = r.Get(ctx, req.NamespacedName, tws); err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, could have been deleted
			logger.Info("TimeWindowScaler not found, may have been deleted", "name", req.Name, "namespace", req.Namespace)
			err = nil // Clear error since this is not a failure
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get TimeWindowScaler", "name", req.Name, "namespace", req.Namespace)
		return ctrl.Result{}, err
	}

	// Initialize clock if not set (for testing)
	if r.Clock == nil {
		r.Clock = engine.RealClock{}
	}

	// Check if the object is being deleted
	if tws.ObjectMeta.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, tws)
	}

	// Add finalizer if not present
	if !containsString(tws.ObjectMeta.Finalizers, timeWindowScalerFinalizer) {
		tws.ObjectMeta.Finalizers = append(tws.ObjectMeta.Finalizers, timeWindowScalerFinalizer)
		if err = r.Update(ctx, tws); err != nil {
			logger.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		// Requeue to continue processing
		return ctrl.Result{Requeue: true}, nil
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

	if err = r.Get(ctx, deploymentKey, deployment); err != nil {
		if apierrors.IsNotFound(err) {
			// Target not found - update status and requeue
			return r.handleMissingTarget(ctx, tws)
		}
		logger.Error(err, "Failed to get target deployment",
			"deployment", deploymentKey.Name,
			"namespace", deploymentKey.Namespace)
		// Set error condition
		r.setErrorCondition(tws, "TargetFetchFailed",
			fmt.Sprintf("Failed to get target deployment %s: %v", deploymentKey.Name, err))
		if statusErr := r.Status().Update(ctx, tws); statusErr != nil {
			logger.Error(statusErr, "Failed to update status after deployment fetch error")
		}
		r.Recorder.Event(tws, corev1.EventTypeWarning, "TargetFetchFailed",
			fmt.Sprintf("Failed to get target deployment %s: %v", deploymentKey.Name, err))
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
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
	var engineInput engine.Input
	engineInput, err = r.buildEngineInput(ctx, tws)
	if err != nil {
		logger.Error(err, "Failed to build engine input")
		r.setErrorCondition(tws, "InvalidConfiguration",
			fmt.Sprintf("Failed to build engine input: %v", err))
		if statusErr := r.Status().Update(ctx, tws); statusErr != nil {
			logger.Error(statusErr, "Failed to update status after build input error")
		}
		r.Recorder.Event(tws, corev1.EventTypeWarning, "InvalidConfiguration",
			fmt.Sprintf("Failed to build engine input: %v", err))
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}
	var engineOutput engine.Output
	engineOutput, err = engine.ComputeEffectiveReplicas(engineInput)
	if err != nil {
		logger.Error(err, "Failed to compute effective replicas")
		r.setErrorCondition(tws, "ComputeFailed",
			fmt.Sprintf("Failed to compute effective replicas: %v", err))
		if statusErr := r.Status().Update(ctx, tws); statusErr != nil {
			logger.Error(statusErr, "Failed to update status after compute error")
		}
		r.Recorder.Event(tws, corev1.EventTypeWarning, "ComputeFailed",
			fmt.Sprintf("Failed to compute effective replicas: %v", err))
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
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
		if err = r.scaleDeployment(ctx, deployment, targetReplicas); err != nil {
			logger.Error(err, "Failed to scale deployment",
				"deployment", deployment.Name,
				"from", currentReplicas,
				"to", targetReplicas)
			r.setErrorCondition(tws, "ScaleFailed",
				fmt.Sprintf("Failed to scale deployment from %d to %d: %v", currentReplicas, targetReplicas, err))
			if statusErr := r.Status().Update(ctx, tws); statusErr != nil {
				logger.Error(statusErr, "Failed to update status after scale error")
			}
			r.Recorder.Event(tws, corev1.EventTypeWarning, "ScaleFailed",
				fmt.Sprintf("Failed to scale deployment from %d to %d: %v", currentReplicas, targetReplicas, err))
			return ctrl.Result{RequeueAfter: 30 * time.Second}, err
		}

		// Emit event and track metrics
		direction := "up"
		eventType := "ScaledUp"
		if targetReplicas < currentReplicas {
			eventType = "ScaledDown"
			direction = "down"
		}
		r.Recorder.Event(tws, corev1.EventTypeNormal, eventType,
			fmt.Sprintf("Scaled from %d to %d replicas (window: %s)",
				currentReplicas, targetReplicas, engineOutput.CurrentWindow))

		// Track scale operation metric
		metrics.ScaleOperationsTotal.WithLabelValues(
			tws.Namespace,
			tws.Name,
			direction,
			engineOutput.CurrentWindow,
		).Inc()

		// Update LastScaleTime when we actually scale
		tws.Status.LastScaleTime = &metav1.Time{Time: r.Clock.Now()}
	}

	// Update effective replicas gauge
	metrics.EffectiveReplicas.WithLabelValues(
		tws.Namespace,
		tws.Name,
		engineOutput.CurrentWindow,
	).Set(float64(targetReplicas))

	// Track window transitions
	previousWindow := tws.Status.CurrentWindow
	if previousWindow != "" && previousWindow != engineOutput.CurrentWindow {
		metrics.WindowTransitionsTotal.WithLabelValues(
			tws.Namespace,
			tws.Name,
			previousWindow,
			engineOutput.CurrentWindow,
		).Inc()
	}

	// Update status
	tws.Status.ObservedGeneration = tws.Generation
	tws.Status.EffectiveReplicas = &engineOutput.EffectiveReplicas
	tws.Status.TargetObservedReplicas = deployment.Spec.Replicas
	tws.Status.CurrentWindow = engineOutput.CurrentWindow
	nextBoundaryTime := metav1.NewTime(engineOutput.NextBoundary)
	tws.Status.NextBoundary = &nextBoundaryTime

	// Handle grace period expiry tracking
	if engineOutput.Reason == "grace-period-active" && tws.Spec.GracePeriodSeconds != nil {
		// Calculate and store grace period expiry time
		if tws.Status.LastScaleTime != nil {
			gracePeriodExpiry := tws.Status.LastScaleTime.Time.Add(time.Duration(*tws.Spec.GracePeriodSeconds) * time.Second)
			expiryTime := metav1.NewTime(gracePeriodExpiry)

			// Emit event if grace period just became active
			if tws.Status.GracePeriodExpiry == nil {
				r.Recorder.Event(tws, corev1.EventTypeNormal, "GracePeriodActive",
					fmt.Sprintf("Grace period active until %s, maintaining %d replicas",
						gracePeriodExpiry.Format(time.RFC3339), engineOutput.EffectiveReplicas))
			}
			tws.Status.GracePeriodExpiry = &expiryTime
		}
	} else {
		// Clear grace period expiry when not in grace period
		if tws.Status.GracePeriodExpiry != nil {
			r.Recorder.Event(tws, corev1.EventTypeNormal, "GracePeriodEnded",
				"Grace period has ended, normal scaling resumed")
		}
		tws.Status.GracePeriodExpiry = nil
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
	if err = r.Status().Update(ctx, tws); err != nil {
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
func (r *TimeWindowScalerReconciler) buildEngineInput(ctx context.Context, tws *kyklosv1alpha1.TimeWindowScaler) (engine.Input, error) {
	logger := log.FromContext(ctx)

	// Validate and convert windows
	windows := make([]engine.WindowSpec, len(tws.Spec.Windows))
	for i, w := range tws.Spec.Windows {
		// Validate time format
		if err := r.validateTimeFormat(w.Start); err != nil {
			errMsg := fmt.Sprintf("Window '%s' has invalid start time '%s': %v", w.Name, w.Start, err)
			logger.Error(err, errMsg)
			r.Recorder.Event(tws, corev1.EventTypeWarning, "InvalidWindow",
				errMsg)
			return engine.Input{}, fmt.Errorf("%s", errMsg)
		}
		if err := r.validateTimeFormat(w.End); err != nil {
			errMsg := fmt.Sprintf("Window '%s' has invalid end time '%s': %v", w.Name, w.End, err)
			logger.Error(err, errMsg)
			r.Recorder.Event(tws, corev1.EventTypeWarning, "InvalidWindow",
				errMsg)
			return engine.Input{}, fmt.Errorf("%s", errMsg)
		}

		// Validate replicas
		if w.Replicas < 0 {
			errMsg := fmt.Sprintf("Window '%s' has invalid replicas %d: must be >= 0", w.Name, w.Replicas)
			logger.Error(nil, errMsg)
			r.Recorder.Event(tws, corev1.EventTypeWarning, "InvalidWindow",
				errMsg)
			return engine.Input{}, fmt.Errorf("%s", errMsg)
		}

		windows[i] = engine.WindowSpec{
			Start:    w.Start,
			End:      w.End,
			Replicas: w.Replicas,
			Name:     w.Name,
			Days:     w.Days,
		}
	}

	// Check if today is a holiday
	isHoliday := false
	previousHolidayState := false // Track state changes for events
	if tws.Spec.HolidayConfigMap != nil && *tws.Spec.HolidayConfigMap != "" {
		holiday, err := r.checkHoliday(ctx, tws.Namespace, *tws.Spec.HolidayConfigMap, tws.Spec.Timezone)
		if err != nil {
			// Log error and emit event - holidays are optional
			log.FromContext(ctx).Error(err, "Failed to check holiday ConfigMap", "configmap", *tws.Spec.HolidayConfigMap)
			r.Recorder.Event(tws, corev1.EventTypeWarning, "HolidayCheckFailed",
				fmt.Sprintf("Failed to check holiday ConfigMap %s: %v", *tws.Spec.HolidayConfigMap, err))
		} else {
			isHoliday = holiday
			// Emit event if holiday state changed
			if isHoliday && !previousHolidayState {
				r.Recorder.Event(tws, corev1.EventTypeNormal, "HolidayDetected",
					fmt.Sprintf("Today is a holiday (mode: %s)", tws.Spec.HolidayMode))
			}
		}
	}

	input := engine.Input{
		Now:             r.Clock.Now(),
		Timezone:        tws.Spec.Timezone,
		Windows:         windows,
		DefaultReplicas: tws.Spec.DefaultReplicas,
		HolidayMode:     tws.Spec.HolidayMode,
		IsHoliday:       isHoliday,
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

	return input, nil
}

// checkHoliday checks if today is a holiday in the ConfigMap
func (r *TimeWindowScalerReconciler) checkHoliday(ctx context.Context, namespace, configMapName, timezone string) (bool, error) {
	// Load timezone
	var err error
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return false, fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}

	// Get current date in the specified timezone
	now := r.Clock.Now().In(loc)
	todayKey := now.Format("2006-01-02") // YYYY-MM-DD format

	// Fetch the ConfigMap
	cm := &corev1.ConfigMap{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      configMapName,
	}, cm); err != nil {
		if apierrors.IsNotFound(err) {
			// ConfigMap doesn't exist - not a holiday
			return false, nil
		}
		return false, fmt.Errorf("failed to get ConfigMap: %w", err)
	}

	// Check if today's date is in the ConfigMap data
	if cm.Data != nil {
		if _, exists := cm.Data[todayKey]; exists {
			return true, nil
		}
	}

	return false, nil
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
	engineInput, err := r.buildEngineInput(ctx, tws)
	if err != nil {
		return ctrl.Result{}, err
	}
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

	if err = r.Status().Update(ctx, tws); err != nil {
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
		Owns(&appsv1.Deployment{}).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(r.findTimeWindowScalersForConfigMap),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Named("timewindowscaler").
		Complete(r)
}

// findTimeWindowScalersForConfigMap finds all TimeWindowScaler resources that reference a ConfigMap
func (r *TimeWindowScalerReconciler) findTimeWindowScalersForConfigMap(ctx context.Context, obj client.Object) []reconcile.Request {
	cm := obj.(*corev1.ConfigMap)
	logger := log.FromContext(ctx)

	// List all TimeWindowScalers in the same namespace
	twsList := &kyklosv1alpha1.TimeWindowScalerList{}
	if err := r.List(ctx, twsList, client.InNamespace(cm.Namespace)); err != nil {
		logger.Error(err, "Failed to list TimeWindowScalers for ConfigMap change", "configmap", cm.Name)
		return nil
	}

	var requests []reconcile.Request
	for _, tws := range twsList.Items {
		// Check if this TWS references the ConfigMap
		if tws.Spec.HolidayConfigMap != nil && *tws.Spec.HolidayConfigMap == cm.Name {
			logger.Info("ConfigMap changed, triggering reconciliation",
				"configmap", cm.Name,
				"timewindowscaler", tws.Name)
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      tws.Name,
					Namespace: tws.Namespace,
				},
			})
		}
	}

	return requests
}

// handleDeletion handles the cleanup when a TimeWindowScaler is being deleted
func (r *TimeWindowScalerReconciler) handleDeletion(ctx context.Context, tws *kyklosv1alpha1.TimeWindowScaler) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if containsString(tws.ObjectMeta.Finalizers, timeWindowScalerFinalizer) {
		// Perform cleanup
		logger.Info("Performing cleanup for TimeWindowScaler",
			"name", tws.Name,
			"namespace", tws.Namespace)

		// Clean up metrics
		metrics.ScaleOperationsTotal.DeleteLabelValues(tws.Namespace, tws.Name)
		metrics.EffectiveReplicas.DeleteLabelValues(tws.Namespace, tws.Name)
		metrics.WindowTransitionsTotal.DeleteLabelValues(tws.Namespace, tws.Name)
		metrics.ReconcileDurationSeconds.DeleteLabelValues(tws.Namespace, tws.Name)

		// Emit deletion event
		r.Recorder.Event(tws, corev1.EventTypeNormal, "Deleting",
			"TimeWindowScaler is being deleted, cleaning up resources")

		// Remove finalizer
		tws.ObjectMeta.Finalizers = removeString(tws.ObjectMeta.Finalizers, timeWindowScalerFinalizer)
		if err := r.Update(ctx, tws); err != nil {
			logger.Error(err, "Failed to remove finalizer")
			return ctrl.Result{}, err
		}

		logger.Info("Successfully cleaned up TimeWindowScaler",
			"name", tws.Name,
			"namespace", tws.Namespace)
	}

	return ctrl.Result{}, nil
}

// setErrorCondition sets an error condition on the TimeWindowScaler
func (r *TimeWindowScalerReconciler) setErrorCondition(tws *kyklosv1alpha1.TimeWindowScaler, reason, message string) {
	errorCondition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		ObservedGeneration: tws.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
	meta.SetStatusCondition(&tws.Status.Conditions, errorCondition)
}

// containsString checks if a string slice contains a string
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// removeString removes a string from a string slice
func removeString(slice []string, s string) []string {
	var result []string
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}

// validateTimeFormat validates HH:MM time format
func (r *TimeWindowScalerReconciler) validateTimeFormat(timeStr string) error {
	_, err := time.Parse("15:04", timeStr)
	if err != nil {
		return fmt.Errorf("invalid time format (expected HH:MM): %w", err)
	}
	return nil
}
