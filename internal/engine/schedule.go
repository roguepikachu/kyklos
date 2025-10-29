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

package engine

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Window represents a parsed time window
type Window struct {
	Start    time.Time // Start time today or tomorrow
	End      time.Time // End time today or tomorrow
	Replicas int32
	Name     string
	Days     []string // Optional day restriction
}

// Input contains the input for computing effective replicas
type Input struct {
	Now              time.Time
	Timezone         string
	Windows          []WindowSpec
	DefaultReplicas  int32
	HolidayMode      string
	IsHoliday        bool
	Pause            bool
	GracePeriodSecs  int32
	LastScaleTime    *time.Time
	CurrentReplicas  int32
}

// WindowSpec is a window specification from the API
type WindowSpec struct {
	Start    string   // HH:MM format
	End      string   // HH:MM format
	Replicas int32
	Name     string
	Days     []string // Optional: ["Monday", "Tuesday"]
}

// Output contains the computed values
type Output struct {
	EffectiveReplicas int32
	NextBoundary      time.Time
	CurrentWindow     string
	Reason            string
}

// ComputeEffectiveReplicas calculates the desired replica count based on time windows
func ComputeEffectiveReplicas(input Input) (Output, error) {
	// Load timezone
	loc, err := time.LoadLocation(input.Timezone)
	if err != nil {
		return Output{}, fmt.Errorf("invalid timezone %s: %w", input.Timezone, err)
	}

	// Get current time in the specified timezone
	nowLocal := input.Now.In(loc)

	// Handle pause mode - compute but don't apply
	if input.Pause {
		out := computeWithoutPause(input, nowLocal, loc)
		out.Reason = "paused"
		return out, nil
	}

	// Handle holiday modes
	if input.IsHoliday {
		switch input.HolidayMode {
		case "treat-as-closed":
			return Output{
				EffectiveReplicas: 0,
				NextBoundary:      getNextDayStart(nowLocal),
				CurrentWindow:     "Holiday-Closed",
				Reason:            "holiday-closed",
			}, nil
		case "treat-as-open":
			// Find max replicas among all windows
			maxReplicas := input.DefaultReplicas
			for _, ws := range input.Windows {
				if ws.Replicas > maxReplicas {
					maxReplicas = ws.Replicas
				}
			}
			return Output{
				EffectiveReplicas: maxReplicas,
				NextBoundary:      getNextDayStart(nowLocal),
				CurrentWindow:     "Holiday-Open",
				Reason:            "holiday-open",
			}, nil
		case "ignore":
			// Continue with normal window processing
		}
	}

	return computeWithoutPause(input, nowLocal, loc), nil
}

func computeWithoutPause(input Input, nowLocal time.Time, loc *time.Location) Output {
	// Parse all windows and find matches
	var activeWindow *Window
	var nextBoundary time.Time
	nextBoundary = getNextDayStart(nowLocal) // Default to tomorrow

	// Process windows in reverse order (last wins)
	for i := len(input.Windows) - 1; i >= 0; i-- {
		ws := input.Windows[i]

		// Check day restriction
		if !isDayMatch(ws.Days, nowLocal) {
			continue
		}

		// Parse window times
		window, err := parseWindow(ws, nowLocal, loc)
		if err != nil {
			continue // Skip invalid windows
		}

		// Check if we're in this window
		if isInWindow(nowLocal, window) {
			if activeWindow == nil {
				activeWindow = window
			}
		}

		// Update next boundary
		boundary := getWindowBoundary(nowLocal, window)
		if boundary.Before(nextBoundary) && boundary.After(nowLocal) {
			nextBoundary = boundary
		}
	}

	// Determine effective replicas and window name
	if activeWindow != nil {
		windowName := activeWindow.Name
		if windowName == "" {
			windowName = fmt.Sprintf("%s-%s",
				activeWindow.Start.Format("15:04"),
				activeWindow.End.Format("15:04"))
		}
		return Output{
			EffectiveReplicas: activeWindow.Replicas,
			NextBoundary:      nextBoundary,
			CurrentWindow:     windowName,
			Reason:            "in-window",
		}
	}

	// No matching window - use default
	return Output{
		EffectiveReplicas: input.DefaultReplicas,
		NextBoundary:      nextBoundary,
		CurrentWindow:     "Default",
		Reason:            "no-matching-window",
	}
}

// parseWindow converts a WindowSpec into a Window with absolute times
func parseWindow(ws WindowSpec, nowLocal time.Time, loc *time.Location) (*Window, error) {
	// Parse start time for today
	startTime, err := parseTimeString(ws.Start, nowLocal, loc)
	if err != nil {
		return nil, err
	}

	// Parse end time for today initially
	endTime, err := parseTimeString(ws.End, nowLocal, loc)
	if err != nil {
		return nil, err
	}

	// Check if this is a cross-midnight window
	// If end hour:minute is less than or equal to start hour:minute, it crosses midnight
	startHour, startMin := startTime.Hour(), startTime.Minute()
	endHour, endMin := endTime.Hour(), endTime.Minute()

	if endHour < startHour || (endHour == startHour && endMin <= startMin) {
		// This is a cross-midnight window
		// If we're currently in the early morning part (before the end time),
		// we need to adjust the start to be yesterday
		if nowLocal.Hour() < endHour || (nowLocal.Hour() == endHour && nowLocal.Minute() < endMin) {
			// We're in the early morning part - start was yesterday
			startTime = startTime.AddDate(0, 0, -1)
		} else {
			// We're in the late night part or outside - end is tomorrow
			endTime = endTime.AddDate(0, 0, 1)
		}
	}

	return &Window{
		Start:    startTime,
		End:      endTime,
		Replicas: ws.Replicas,
		Name:     ws.Name,
		Days:     ws.Days,
	}, nil
}

// parseTimeString parses HH:MM format into a time.Time for today
func parseTimeString(timeStr string, nowLocal time.Time, loc *time.Location) (time.Time, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return time.Time{}, fmt.Errorf("invalid hour: %s", parts[0])
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return time.Time{}, fmt.Errorf("invalid minute: %s", parts[1])
	}

	// Create time for today at the specified hour:minute
	result := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(),
		hour, minute, 0, 0, loc)

	return result, nil
}

// isInWindow checks if the current time is within a window (start inclusive, end exclusive)
func isInWindow(now time.Time, window *Window) bool {
	// For cross-midnight windows, we need to check if we're:
	// - Between start and midnight (same day as start), OR
	// - Between midnight and end (next day)

	if window.End.Before(window.Start) || window.End.Day() != window.Start.Day() {
		// This is a cross-midnight window
		// Check if we're in the late-night part (after start, before midnight)
		if now.Day() == window.Start.Day() &&
		   (now.Equal(window.Start) || now.After(window.Start)) {
			return true
		}
		// Check if we're in the early-morning part (after midnight, before end)
		if now.Day() == window.End.Day() && now.Before(window.End) {
			return true
		}
		return false
	}

	// Normal window (doesn't cross midnight)
	return (now.Equal(window.Start) || now.After(window.Start)) && now.Before(window.End)
}

// isDayMatch checks if the current day matches the window's day restrictions
func isDayMatch(days []string, now time.Time) bool {
	if len(days) == 0 {
		return true // No restriction
	}

	todayName := now.Weekday().String()
	for _, day := range days {
		if strings.EqualFold(day, todayName) {
			return true
		}
	}

	return false
}

// getWindowBoundary returns the next start or end time for a window
func getWindowBoundary(now time.Time, window *Window) time.Time {
	// For cross-midnight windows, we need special handling
	if window.End.Day() != window.Start.Day() {
		// Cross-midnight window
		if isInWindow(now, window) {
			// We're in the window, next boundary is the end
			return window.End
		}
		// We're outside - find the next start
		// If we're before today's start time
		if now.Before(window.Start) {
			return window.Start
		}
		// We're after the window end, next occurrence is tonight
		nextStart := time.Date(now.Year(), now.Month(), now.Day(),
			window.Start.Hour(), window.Start.Minute(), 0, 0, now.Location())
		if nextStart.Before(now) || nextStart.Equal(now) {
			nextStart = nextStart.AddDate(0, 0, 1)
		}
		return nextStart
	}

	// Normal (non-cross-midnight) window
	if now.Before(window.Start) {
		return window.Start
	}
	if now.Before(window.End) {
		return window.End
	}
	// We're after the window - next boundary is tomorrow's start
	return window.Start.AddDate(0, 0, 1)
}

// getNextDayStart returns midnight tomorrow in the same timezone
func getNextDayStart(now time.Time) time.Time {
	tomorrow := now.AddDate(0, 0, 1)
	return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
		0, 0, 0, 0, now.Location())
}

// ComputeNextBoundary calculates the next time when scaling might change
func ComputeNextBoundary(input Input) (time.Time, error) {
	output, err := ComputeEffectiveReplicas(input)
	if err != nil {
		return time.Time{}, err
	}
	return output.NextBoundary, nil
}