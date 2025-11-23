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

// Package metrics provides Prometheus metrics for Kyklos
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// ScaleOperationsTotal tracks the total number of scaling operations
	ScaleOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kyklos_scale_operations_total",
			Help: "Total number of scaling operations performed by Kyklos",
		},
		[]string{"namespace", "name", "direction", "window"},
	)

	// EffectiveReplicas tracks the current effective replica count
	EffectiveReplicas = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kyklos_effective_replicas",
			Help: "Current effective replica count computed by Kyklos",
		},
		[]string{"namespace", "name", "window"},
	)

	// WindowTransitionsTotal tracks window transitions
	WindowTransitionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kyklos_window_transitions_total",
			Help: "Total number of window transitions",
		},
		[]string{"namespace", "name", "from_window", "to_window"},
	)

	// ReconcileDurationSeconds tracks reconciliation duration
	ReconcileDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kyklos_reconcile_duration_seconds",
			Help:    "Time taken for reconciliation in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"namespace", "name", "result"},
	)
)

func init() {
	// Register metrics with controller-runtime's metrics registry
	metrics.Registry.MustRegister(
		ScaleOperationsTotal,
		EffectiveReplicas,
		WindowTransitionsTotal,
		ReconcileDurationSeconds,
	)
}
