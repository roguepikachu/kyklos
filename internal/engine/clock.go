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

// Package engine provides pure time-based scheduling logic without Kubernetes dependencies
package engine

import (
	"time"
)

// Clock provides an interface for time operations to support testing
type Clock interface {
	Now() time.Time
}

// RealClock uses actual system time
type RealClock struct{}

// Now returns the current time
func (r RealClock) Now() time.Time {
	return time.Now()
}

// FakeClock returns a fixed time for testing
type FakeClock struct {
	Time time.Time
}

// Now returns the fixed time
func (f FakeClock) Now() time.Time {
	return f.Time
}
