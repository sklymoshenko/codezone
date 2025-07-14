package executor

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "nanoseconds under 1 microsecond",
			duration: 500 * time.Nanosecond,
			expected: "500ns",
		},
		{
			name:     "single nanosecond",
			duration: 1 * time.Nanosecond,
			expected: "1ns",
		},
		{
			name:     "microseconds under 1 millisecond",
			duration: 1500 * time.Nanosecond, // 1.5 microseconds
			expected: "1.5μs",
		},
		{
			name:     "microseconds with precision",
			duration: 1234 * time.Nanosecond, // 1.234 microseconds
			expected: "1.23μs",               // Should round to 3 significant digits
		},
		{
			name:     "exactly 1 microsecond",
			duration: 1 * time.Microsecond,
			expected: "1μs",
		},
		{
			name:     "milliseconds under 1 second",
			duration: 1500 * time.Microsecond, // 1.5 milliseconds
			expected: "1.5ms",
		},
		{
			name:     "milliseconds with precision",
			duration: 1814595 * time.Nanosecond, // 1.814595 milliseconds
			expected: "1.81ms",                  // Should format to 3 significant digits
		},
		{
			name:     "exactly 1 millisecond",
			duration: 1 * time.Millisecond,
			expected: "1ms",
		},
		{
			name:     "seconds with decimal",
			duration: 1500 * time.Millisecond, // 1.5 seconds
			expected: "1.5s",
		},
		{
			name:     "seconds with precision",
			duration: 2347 * time.Millisecond, // 2.347 seconds
			expected: "2.35s",                 // Should round to 3 significant digits
		},
		{
			name:     "exactly 1 second",
			duration: 1 * time.Second,
			expected: "1s",
		},
		{
			name:     "large duration in seconds",
			duration: 65 * time.Second, // 65 seconds
			expected: "65s",
		},
		{
			name:     "very small microseconds",
			duration: 100 * time.Nanosecond, // 0.1 microseconds
			expected: "100ns",               // Actually still in nanosecond range
		},
		{
			name:     "very small milliseconds",
			duration: 100 * time.Microsecond, // 0.1 milliseconds
			expected: "100μs",                // Actually still in microsecond range
		},
		{
			name:     "very small seconds",
			duration: 100 * time.Millisecond, // 0.1 seconds
			expected: "100ms",                // Actually still in millisecond range
		},
		{
			name:     "zero duration",
			duration: 0,
			expected: "0ns",
		},
		{
			name:     "large nanoseconds",
			duration: 999 * time.Nanosecond,
			expected: "999ns",
		},
		{
			name:     "boundary microsecond",
			duration: 999500 * time.Nanosecond, // 999.5 microseconds
			expected: "1e+03μs",                // Go's %.3g format uses scientific notation
		},
		{
			name:     "boundary millisecond",
			duration: 999500 * time.Microsecond, // 999.5 milliseconds
			expected: "1e+03ms",                 // Go's %.3g format uses scientific notation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Fatalf("formatDuration(%v) = %s, expected %s", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestFormatDuration_Precision(t *testing.T) {
	t.Run("should not exceed 3 significant digits", func(t *testing.T) {
		// Test that we don't get overly precise results
		duration := 1814595123 * time.Nanosecond // Very precise nanoseconds
		result := formatDuration(duration)
		expected := "1.81s" // Should be rounded to reasonable precision
		if result != expected {
			t.Fatalf("formatDuration(%v) = %s, expected %s", duration, result, expected)
		}
	})

	t.Run("should handle edge case between units", func(t *testing.T) {
		// Test boundary between microseconds and milliseconds
		duration := 999950 * time.Nanosecond // 999.95 microseconds
		result := formatDuration(duration)
		// Go's %.3g format will use scientific notation for this
		expected := "1e+03μs"
		if result != expected {
			t.Fatalf("formatDuration(%v) = %s, expected %s", duration, result, expected)
		}
	})
}

func BenchmarkFormatDuration(b *testing.B) {
	durations := []time.Duration{
		100 * time.Nanosecond,
		1500 * time.Nanosecond,
		1500 * time.Microsecond,
		1500 * time.Millisecond,
	}

	for _, d := range durations {
		b.Run(d.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				formatDuration(d)
			}
		})
	}
}
