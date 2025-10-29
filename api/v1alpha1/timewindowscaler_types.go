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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TimeWindowScalerSpec defines the desired state of TimeWindowScaler
type TimeWindowScalerSpec struct {
	// TargetRef identifies the Deployment to scale
	// +kubebuilder:validation:Required
	TargetRef TargetRef `json:"targetRef"`

	// DefaultReplicas is the replica count when no windows match
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1
	DefaultReplicas int32 `json:"defaultReplicas"`

	// Timezone for evaluating time windows (IANA timezone)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[A-Za-z]+(/[A-Za-z_]+)?(/[A-Za-z_]+)?$`
	// +kubebuilder:example="America/New_York"
	Timezone string `json:"timezone"`

	// Windows define time-based scaling rules
	// +optional
	Windows []TimeWindow `json:"windows,omitempty"`

	// HolidayMode determines how holidays affect scaling
	// +kubebuilder:validation:Enum=ignore;treat-as-closed;treat-as-open
	// +kubebuilder:default="ignore"
	// +optional
	HolidayMode string `json:"holidayMode,omitempty"`

	// HolidayConfigMap references a ConfigMap with holiday dates
	// +optional
	HolidayConfigMap *string `json:"holidayConfigMap,omitempty"`

	// GracePeriodSeconds for scale-down operations
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=3600
	// +kubebuilder:default=300
	// +optional
	GracePeriodSeconds *int32 `json:"gracePeriodSeconds,omitempty"`

	// Pause disables all scaling operations
	// +kubebuilder:default=false
	// +optional
	Pause bool `json:"pause,omitempty"`
}

// TargetRef identifies the target workload
type TargetRef struct {
	// Name of the Deployment
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace of the Deployment (defaults to TWS namespace)
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// TimeWindow defines a time-based scaling rule
type TimeWindow struct {
	// Start time in HH:MM format (24-hour)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$`
	Start string `json:"start"`

	// End time in HH:MM format (24-hour)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$`
	End string `json:"end"`

	// Replicas to maintain during this window
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Required
	Replicas int32 `json:"replicas"`

	// Days when this window is active
	// +optional
	Days []string `json:"days,omitempty"`

	// Name for this window (used in labels)
	// +optional
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name,omitempty"`
}

// TimeWindowScalerStatus defines the observed state of TimeWindowScaler.
type TimeWindowScalerStatus struct {
	// ObservedGeneration tracks the generation of the spec
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// EffectiveReplicas is the computed desired replica count
	// +optional
	EffectiveReplicas *int32 `json:"effectiveReplicas,omitempty"`

	// TargetObservedReplicas is the observed replica count on the target
	// +optional
	TargetObservedReplicas *int32 `json:"targetObservedReplicas,omitempty"`

	// CurrentWindow indicates the active time window
	// +optional
	CurrentWindow string `json:"currentWindow,omitempty"`

	// NextBoundary is the next time a scaling action might occur
	// +optional
	NextBoundary *metav1.Time `json:"nextBoundary,omitempty"`

	// LastScaleTime is when the last scaling action occurred
	// +optional
	LastScaleTime *metav1.Time `json:"lastScaleTime,omitempty"`

	// GracePeriodExpiry indicates when grace period ends
	// +optional
	GracePeriodExpiry *metav1.Time `json:"gracePeriodExpiry,omitempty"`

	// Conditions represent the latest observations of the resource state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=tws
// +kubebuilder:printcolumn:name="Target",type="string",JSONPath=".spec.targetRef.name"
// +kubebuilder:printcolumn:name="Default",type="integer",JSONPath=".spec.defaultReplicas"
// +kubebuilder:printcolumn:name="Effective",type="integer",JSONPath=".status.effectiveReplicas"
// +kubebuilder:printcolumn:name="Window",type="string",JSONPath=".status.currentWindow"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// TimeWindowScaler is the Schema for the timewindowscalers API
type TimeWindowScaler struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of TimeWindowScaler
	// +required
	Spec TimeWindowScalerSpec `json:"spec"`

	// status defines the observed state of TimeWindowScaler
	// +optional
	Status TimeWindowScalerStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// TimeWindowScalerList contains a list of TimeWindowScaler
type TimeWindowScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TimeWindowScaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TimeWindowScaler{}, &TimeWindowScalerList{})
}
