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
			deploymentName = "test-deployment"
			twsName = "test-tws"

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
			// Clean up
			if tws != nil {
				_ = k8sClient.Delete(ctx, tws)
			}
			if deployment != nil {
				_ = k8sClient.Delete(ctx, deployment)
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

			// Reconcile
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      twsName,
					Namespace: namespace,
				},
			}

			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeNumerically(">", 0))

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

			// Reconcile
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      twsName,
					Namespace: namespace,
				},
			}

			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeNumerically(">", 0))

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

			// Reconcile
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      twsName,
					Namespace: namespace,
				},
			}

			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeNumerically(">", 0))

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

			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(5 * time.Minute))

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
	})
})

func ptr(i int32) *int32 {
	return &i
}