/*
Copyright 2025 roguepikachu.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kyklosv1alpha1 "github.com/roguepikachu/kyklos/api/v1alpha1"
	"github.com/roguepikachu/kyklos/internal/engine"
)

var _ = Describe("TimeWindowScaler Controller", func() {
	const (
		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	var testCounter int

	Context("When scaling a Deployment based on time windows", func() {
		var (
			ctx            context.Context
			namespace      string
			deploymentName string
			twsName        string
			deployment     *appsv1.Deployment
			tws            *kyklosv1alpha1.TimeWindowScaler
			reconciler     *TimeWindowScalerReconciler
		)

		BeforeEach(func() {
			ctx = context.Background()
			namespace = "default"
			testCounter++
			deploymentName = fmt.Sprintf("test-deployment-%d-%d", time.Now().Unix(), testCounter)
			twsName = fmt.Sprintf("test-tws-%d-%d", time.Now().Unix(), testCounter)

			// Create a test deployment
			deployment = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deploymentName,
					Namespace: namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr(int32(1)),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "test",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "nginx:latest",
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, deployment)).To(Succeed())

			// Set up the reconciler with a fake clock
			reconciler = &TimeWindowScalerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: record.NewFakeRecorder(100),
				Clock:    engine.FakeClock{Time: time.Date(2025, 3, 10, 10, 0, 0, 0, time.UTC)}, // Monday 10:00 UTC
			}
		})

		AfterEach(func() {
			// Clean up with proper error handling
			if tws != nil {
				// First remove finalizer if present
				currentTWS := &kyklosv1alpha1.TimeWindowScaler{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: tws.Name, Namespace: tws.Namespace}, currentTWS); err == nil {
					currentTWS.Finalizers = []string{}
					_ = k8sClient.Update(ctx, currentTWS)
				}
				_ = k8sClient.Delete(ctx, tws)
				tws = nil
			}
			if deployment != nil {
				_ = k8sClient.Delete(ctx, deployment)
				deployment = nil
			}
		})

		It("should scale up during business hours", func() {
			// Create a TimeWindowScaler with a business hours window
			tws = &kyklosv1alpha1.TimeWindowScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      twsName,
					Namespace: namespace,
				},
				Spec: kyklosv1alpha1.TimeWindowScalerSpec{
					TargetRef: kyklosv1alpha1.TargetRef{
						Name:      deploymentName,
						Namespace: namespace,
					},
					DefaultReplicas: 1,
					Timezone:        "UTC",
					Windows: []kyklosv1alpha1.TimeWindow{
						{
							Start:    "09:00",
							End:      "17:00",
							Replicas: 5,
							Name:     "BusinessHours",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, tws)).To(Succeed())

			// Reconcile twice - first adds finalizer, second does actual work
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      twsName,
					Namespace: namespace,
				},
			}

			// First reconcile adds finalizer
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			if result.Requeue || result.RequeueAfter > 0 {
				// Second reconcile does the actual scaling
				_, err = reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
			}

			// Check that the deployment was scaled
			updatedDeployment := &appsv1.Deployment{}
			Eventually(func() int32 {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      deploymentName,
					Namespace: namespace,
				}, updatedDeployment)
				if err != nil {
					return -1
				}
				if updatedDeployment.Spec.Replicas == nil {
					return -1
				}
				return *updatedDeployment.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(5)))

			// Check that the TWS status was updated
			updatedTWS := &kyklosv1alpha1.TimeWindowScaler{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      twsName,
					Namespace: namespace,
				}, updatedTWS)
				if err != nil {
					return false
				}
				return updatedTWS.Status.EffectiveReplicas != nil &&
					*updatedTWS.Status.EffectiveReplicas == 5 &&
					updatedTWS.Status.CurrentWindow == "BusinessHours"
			}, timeout, interval).Should(BeTrue())

			// Check Ready condition
			Expect(meta.IsStatusConditionTrue(updatedTWS.Status.Conditions, "Ready")).To(BeTrue())
		})

		It("should use default replicas outside of windows", func() {
			// Set clock to outside business hours
			reconciler.Clock = engine.FakeClock{Time: time.Date(2025, 3, 10, 18, 0, 0, 0, time.UTC)} // Monday 18:00 UTC

			// Create a TimeWindowScaler with a business hours window
			tws = &kyklosv1alpha1.TimeWindowScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      twsName,
					Namespace: namespace,
				},
				Spec: kyklosv1alpha1.TimeWindowScalerSpec{
					TargetRef: kyklosv1alpha1.TargetRef{
						Name:      deploymentName,
						Namespace: namespace,
					},
					DefaultReplicas: 2,
					Timezone:        "UTC",
					Windows: []kyklosv1alpha1.TimeWindow{
						{
							Start:    "09:00",
							End:      "17:00",
							Replicas: 5,
							Name:     "BusinessHours",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, tws)).To(Succeed())

			// Reconcile twice - first adds finalizer, second does actual work
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      twsName,
					Namespace: namespace,
				},
			}

			// First reconcile adds finalizer
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			if result.Requeue || result.RequeueAfter > 0 {
				// Second reconcile does the actual scaling
				_, err = reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
			}

			// Check that the deployment uses default replicas
			updatedDeployment := &appsv1.Deployment{}
			Eventually(func() int32 {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      deploymentName,
					Namespace: namespace,
				}, updatedDeployment)
				if err != nil {
					return -1
				}
				if updatedDeployment.Spec.Replicas == nil {
					return -1
				}
				return *updatedDeployment.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(2)))

			// Check status
			updatedTWS := &kyklosv1alpha1.TimeWindowScaler{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      twsName,
					Namespace: namespace,
				}, updatedTWS)
				if err != nil {
					return false
				}
				return updatedTWS.Status.EffectiveReplicas != nil &&
					*updatedTWS.Status.EffectiveReplicas == 2 &&
					updatedTWS.Status.CurrentWindow == "Default"
			}, timeout, interval).Should(BeTrue())
		})

		It("should not scale when paused", func() {
			// Create a paused TimeWindowScaler
			tws = &kyklosv1alpha1.TimeWindowScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      twsName,
					Namespace: namespace,
				},
				Spec: kyklosv1alpha1.TimeWindowScalerSpec{
					TargetRef: kyklosv1alpha1.TargetRef{
						Name:      deploymentName,
						Namespace: namespace,
					},
					DefaultReplicas: 1,
					Timezone:        "UTC",
					Pause:           true,
					Windows: []kyklosv1alpha1.TimeWindow{
						{
							Start:    "09:00",
							End:      "17:00",
							Replicas: 5,
							Name:     "BusinessHours",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, tws)).To(Succeed())

			// Reconcile twice - first adds finalizer, second does actual work
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      twsName,
					Namespace: namespace,
				},
			}

			// First reconcile adds finalizer
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			if result.Requeue || result.RequeueAfter > 0 {
				// Second reconcile does the actual scaling
				_, err = reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
			}

			// Check that the deployment was NOT scaled (stays at 1)
			updatedDeployment := &appsv1.Deployment{}
			Consistently(func() int32 {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      deploymentName,
					Namespace: namespace,
				}, updatedDeployment)
				if err != nil {
					return -1
				}
				if updatedDeployment.Spec.Replicas == nil {
					return -1
				}
				return *updatedDeployment.Spec.Replicas
			}, time.Second*2, interval).Should(Equal(int32(1)))

			// Check status shows paused
			updatedTWS := &kyklosv1alpha1.TimeWindowScaler{}
			Eventually(func() string {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      twsName,
					Namespace: namespace,
				}, updatedTWS)
				if err != nil {
					return ""
				}
				for _, cond := range updatedTWS.Status.Conditions {
					if cond.Type == "Ready" {
						return cond.Reason
					}
				}
				return ""
			}, timeout, interval).Should(Equal("Paused"))
		})

		It("should handle missing target gracefully", func() {
			// Create a TimeWindowScaler pointing to non-existent deployment
			tws = &kyklosv1alpha1.TimeWindowScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      twsName,
					Namespace: namespace,
				},
				Spec: kyklosv1alpha1.TimeWindowScalerSpec{
					TargetRef: kyklosv1alpha1.TargetRef{
						Name:      "non-existent-deployment",
						Namespace: namespace,
					},
					DefaultReplicas: 1,
					Timezone:        "UTC",
				},
			}
			Expect(k8sClient.Create(ctx, tws)).To(Succeed())

			// Reconcile
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      twsName,
					Namespace: namespace,
				},
			}

			// First reconcile adds finalizer
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			if result.Requeue || result.RequeueAfter > 0 {
				// Second reconcile does the actual work
				_, err = reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
			}
			// Should requeue after 5 minutes to check if target appears

			// Check status shows target not found
			updatedTWS := &kyklosv1alpha1.TimeWindowScaler{}
			Eventually(func() string {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      twsName,
					Namespace: namespace,
				}, updatedTWS)
				if err != nil {
					return ""
				}
				for _, cond := range updatedTWS.Status.Conditions {
					if cond.Type == "Ready" {
						return cond.Reason
					}
				}
				return ""
			}, timeout, interval).Should(Equal("TargetNotFound"))
		})

		It("Should handle holiday ConfigMap correctly", func() {
			// Use unique names for this test
			holidayDeploymentName := "holiday-test-deployment"
			holidayTWSName := "holiday-test-tws"

			// Create a holiday ConfigMap
			holidayConfigMapName := "test-holidays"
			holidayConfigMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      holidayConfigMapName,
					Namespace: namespace,
				},
				Data: map[string]string{
					"2025-01-01": "New Year's Day",
					"2025-12-25": "Christmas",
				},
			}
			Expect(k8sClient.Create(ctx, holidayConfigMap)).To(Succeed())

			// Create test deployment with 1 replica
			holidayDeployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      holidayDeploymentName,
					Namespace: namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr(1),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": holidayDeploymentName,
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": holidayDeploymentName,
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "nginx:latest",
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, holidayDeployment)).To(Succeed())

			// Create TimeWindowScaler with holiday mode
			holidayTWS := &kyklosv1alpha1.TimeWindowScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      holidayTWSName,
					Namespace: namespace,
				},
				Spec: kyklosv1alpha1.TimeWindowScalerSpec{
					TargetRef: kyklosv1alpha1.TargetRef{
						Name: holidayDeploymentName,
					},
					Timezone:         "UTC",
					DefaultReplicas:  1,
					HolidayMode:      "treat-as-closed",
					HolidayConfigMap: &holidayConfigMapName,
					Windows: []kyklosv1alpha1.TimeWindow{
						{
							Start:    "09:00",
							End:      "17:00",
							Replicas: 5,
							Name:     "business-hours",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, holidayTWS)).To(Succeed())

			// Create reconciler with fake clock set to a holiday
			fakeClock := &engine.FakeClock{
				Time: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC), // New Year's Day at 10 AM
			}
			reconciler = &TimeWindowScalerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: record.NewFakeRecorder(10),
				Clock:    fakeClock,
			}

			// Reconcile
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      holidayTWSName,
					Namespace: namespace,
				},
			}

			// First reconcile adds finalizer
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			if result.Requeue || result.RequeueAfter > 0 {
				// Second reconcile does the actual work
				_, err = reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
			}
			// Requeue is expected for time-based scaling

			// Check deployment should be scaled to 0 (holiday mode treat-as-closed)
			Eventually(func() int32 {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      holidayDeploymentName,
					Namespace: namespace,
				}, holidayDeployment)
				if err != nil {
					return -1
				}
				return *holidayDeployment.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(0)))

			// Now test with non-holiday date
			fakeClock.Time = time.Date(2025, 3, 10, 10, 0, 0, 0, time.UTC) // Regular Monday at 10 AM
			_, err = reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			// Check deployment should be scaled to 5 (within business hours)
			Eventually(func() int32 {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      holidayDeploymentName,
					Namespace: namespace,
				}, holidayDeployment)
				if err != nil {
					return -1
				}
				return *holidayDeployment.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(5)))
		})

		It("Should handle grace period for scale-down operations", func() {
			// Create a deployment with 1 replica so the first reconcile triggers a scale-up
			graceDeploymentName := "grace-test-deployment"
			graceTWSName := "grace-test-tws"

			graceDeployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      graceDeploymentName,
					Namespace: namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr(1),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "grace-test",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "grace-test",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "nginx:latest",
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, graceDeployment)).To(Succeed())

			// Use 300s grace period and scale near the end of the window
			// so the grace period extends past the window boundary
			fakeClock := &engine.FakeClock{
				Time: time.Date(2025, 3, 10, 16, 59, 0, 0, time.UTC), // 16:59 - just before window ends
			}

			gracePeriodSeconds := int32(300) // 5 minutes
			graceTWS := &kyklosv1alpha1.TimeWindowScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      graceTWSName,
					Namespace: namespace,
				},
				Spec: kyklosv1alpha1.TimeWindowScalerSpec{
					TargetRef: kyklosv1alpha1.TargetRef{
						Name: graceDeploymentName,
					},
					Timezone:           "UTC",
					DefaultReplicas:    1,
					GracePeriodSeconds: &gracePeriodSeconds,
					Windows: []kyklosv1alpha1.TimeWindow{
						{
							Start:    "09:00",
							End:      "17:00",
							Replicas: 5,
							Name:     "business-hours",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, graceTWS)).To(Succeed())

			// Create reconciler with fake clock
			reconciler = &TimeWindowScalerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: record.NewFakeRecorder(10),
				Clock:    fakeClock,
			}

			// First reconciliation at 16:59 (inside window) - scales from 1 to 5
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      graceTWSName,
					Namespace: namespace,
				},
			}

			// First reconcile adds finalizer
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			if result.Requeue || result.RequeueAfter > 0 {
				// Second reconcile does the actual work
				_, err = reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
			}

			// Check deployment is at 5 replicas
			Eventually(func() int32 {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      graceDeploymentName,
					Namespace: namespace,
				}, graceDeployment)
				if err != nil {
					return -1
				}
				return *graceDeployment.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(5)))

			// Verify LastScaleTime was set (at 16:59)
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      graceTWSName,
					Namespace: namespace,
				}, graceTWS)
				if err != nil {
					return false
				}
				return graceTWS.Status.LastScaleTime != nil
			}, timeout, interval).Should(BeTrue(), "LastScaleTime should be set after scaling")

			// Move time to just after window ends: 17:01
			// Grace period: LastScaleTime(16:59) + 300s = 17:04 → 17:01 is before 17:04 → active
			fakeClock.Time = time.Date(2025, 3, 10, 17, 1, 0, 0, time.UTC)

			// Reconcile again - should NOT scale down due to grace period
			result, err = reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			// Verify deployment stays at 5 replicas (grace period active)
			Consistently(func() int32 {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      graceDeploymentName,
					Namespace: namespace,
				}, graceDeployment)
				if err != nil {
					return -1
				}
				return *graceDeployment.Spec.Replicas
			}, 2*time.Second, interval).Should(Equal(int32(5)))

			// Check status shows grace period is active
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      graceTWSName,
				Namespace: namespace,
			}, graceTWS)).To(Succeed())
			Expect(graceTWS.Status.GracePeriodExpiry).NotTo(BeNil())

			// Move time past grace period: 17:06
			// Grace period: LastScaleTime(16:59) + 300s = 17:04 → 17:06 is after 17:04 → expired
			fakeClock.Time = time.Date(2025, 3, 10, 17, 6, 0, 0, time.UTC)

			// Reconcile again - should now scale down
			result, err = reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			// Check deployment scales down to 1
			Eventually(func() int32 {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      graceDeploymentName,
					Namespace: namespace,
				}, graceDeployment)
				if err != nil {
					return -1
				}
				return *graceDeployment.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(1)))

			// Grace period expiry should be cleared
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      graceTWSName,
				Namespace: namespace,
			}, graceTWS)).To(Succeed())
			Expect(graceTWS.Status.GracePeriodExpiry).To(BeNil())
		})

		It("Should handle cross-midnight windows correctly", func() {
			crossMidnightDeploymentName := "crossmidnight-test-deployment"
			crossMidnightTWSName := "crossmidnight-test-tws"

			crossMidnightDeployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      crossMidnightDeploymentName,
					Namespace: namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr(1),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "crossmidnight-test",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "crossmidnight-test",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "nginx:latest",
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, crossMidnightDeployment)).To(Succeed())

			// Create TWS with cross-midnight window: 22:00 - 06:00
			crossMidnightTWS := &kyklosv1alpha1.TimeWindowScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      crossMidnightTWSName,
					Namespace: namespace,
				},
				Spec: kyklosv1alpha1.TimeWindowScalerSpec{
					TargetRef: kyklosv1alpha1.TargetRef{
						Name: crossMidnightDeploymentName,
					},
					Timezone:        "UTC",
					DefaultReplicas: 1,
					Windows: []kyklosv1alpha1.TimeWindow{
						{
							Start:    "22:00",
							End:      "06:00",
							Replicas: 8,
							Name:     "night-shift",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, crossMidnightTWS)).To(Succeed())

			// Test 1: During the late-night part (23:00) - should scale to 8
			fakeClock := &engine.FakeClock{
				Time: time.Date(2025, 3, 10, 23, 0, 0, 0, time.UTC), // Monday 23:00
			}
			reconciler = &TimeWindowScalerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: record.NewFakeRecorder(10),
				Clock:    fakeClock,
			}

			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      crossMidnightTWSName,
					Namespace: namespace,
				},
			}

			// First reconcile adds finalizer
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			if result.Requeue || result.RequeueAfter > 0 {
				_, err = reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
			}

			// Verify deployment scales to 8
			Eventually(func() int32 {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      crossMidnightDeploymentName,
					Namespace: namespace,
				}, crossMidnightDeployment)
				if err != nil {
					return -1
				}
				return *crossMidnightDeployment.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(8)))

			// Verify window name
			updatedTWS := &kyklosv1alpha1.TimeWindowScaler{}
			Eventually(func() string {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      crossMidnightTWSName,
					Namespace: namespace,
				}, updatedTWS)
				if err != nil {
					return ""
				}
				return updatedTWS.Status.CurrentWindow
			}, timeout, interval).Should(Equal("night-shift"))

			// Test 2: During the early-morning part (03:00) - should still be in window
			fakeClock.Time = time.Date(2025, 3, 11, 3, 0, 0, 0, time.UTC) // Tuesday 03:00
			_, err = reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			// Deployment should stay at 8
			Eventually(func() int32 {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      crossMidnightDeploymentName,
					Namespace: namespace,
				}, crossMidnightDeployment)
				if err != nil {
					return -1
				}
				return *crossMidnightDeployment.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(8)))

			// Test 3: Outside the window (10:00) - should scale to default
			fakeClock.Time = time.Date(2025, 3, 11, 10, 0, 0, 0, time.UTC) // Tuesday 10:00
			_, err = reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			// Deployment should scale to 1 (default)
			Eventually(func() int32 {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      crossMidnightDeploymentName,
					Namespace: namespace,
				}, crossMidnightDeployment)
				if err != nil {
					return -1
				}
				return *crossMidnightDeployment.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(1)))

			// Verify default window
			Eventually(func() string {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      crossMidnightTWSName,
					Namespace: namespace,
				}, updatedTWS)
				if err != nil {
					return ""
				}
				return updatedTWS.Status.CurrentWindow
			}, timeout, interval).Should(Equal("Default"))
		})

		It("Should not apply grace period for scale-up operations", func() {
			// Create a deployment with 1 replica
			scaleUpDeploymentName := "scaleup-test-deployment"
			scaleUpTWSName := "scaleup-test-tws"

			scaleUpDeployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      scaleUpDeploymentName,
					Namespace: namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr(1),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "scaleup-test",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "scaleup-test",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "nginx:latest",
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, scaleUpDeployment)).To(Succeed())

			// Create TimeWindowScaler with grace period
			gracePeriodSeconds := int32(300) // 5 minutes
			scaleUpTWS := &kyklosv1alpha1.TimeWindowScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      scaleUpTWSName,
					Namespace: namespace,
				},
				Spec: kyklosv1alpha1.TimeWindowScalerSpec{
					TargetRef: kyklosv1alpha1.TargetRef{
						Name: scaleUpDeploymentName,
					},
					Timezone:           "UTC",
					DefaultReplicas:    1,
					GracePeriodSeconds: &gracePeriodSeconds,
					Windows: []kyklosv1alpha1.TimeWindow{
						{
							Start:    "09:00",
							End:      "17:00",
							Replicas: 10,
							Name:     "peak-hours",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, scaleUpTWS)).To(Succeed())

			// Create reconciler with fake clock set to window start
			fakeClock := &engine.FakeClock{
				Time: time.Date(2025, 3, 10, 9, 1, 0, 0, time.UTC), // 09:01 - just after window starts
			}
			reconciler = &TimeWindowScalerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: record.NewFakeRecorder(10),
				Clock:    fakeClock,
			}

			// Reconcile
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      scaleUpTWSName,
					Namespace: namespace,
				},
			}

			// First reconcile adds finalizer
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			if result.Requeue || result.RequeueAfter > 0 {
				// Second reconcile does the actual work
				_, err = reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
			}

			// Check deployment scales up immediately to 10 (no grace period for scale-up)
			Eventually(func() int32 {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      scaleUpDeploymentName,
					Namespace: namespace,
				}, scaleUpDeployment)
				if err != nil {
					return -1
				}
				return *scaleUpDeployment.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(10)))

			// Grace period expiry should not be set for scale-up
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      scaleUpTWSName,
				Namespace: namespace,
			}, scaleUpTWS)).To(Succeed())
			Expect(scaleUpTWS.Status.GracePeriodExpiry).To(BeNil())
		})
	})
})

func ptr(i int32) *int32 {
	return &i
}
