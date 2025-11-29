package models

import (
	"math"
	"testing"
)

func TestLocation_DistanceTo(t *testing.T) {
	tests := []struct {
		name     string
		from     Location
		to       Location
		expected float64
		delta    float64
	}{
		{
			name:     "Same location",
			from:     Location{Latitude: 13.7563, Longitude: 100.5018},
			to:       Location{Latitude: 13.7563, Longitude: 100.5018},
			expected: 0,
			delta:    0.001,
		},
		{
			name:     "Bangkok to Chiang Mai (approximately 580km)",
			from:     Location{Latitude: 13.7563, Longitude: 100.5018},
			to:       Location{Latitude: 18.7883, Longitude: 98.9853},
			expected: 580,
			delta:    20, // Allow 20km tolerance
		},
		{
			name:     "Short distance (about 1km)",
			from:     Location{Latitude: 13.7563, Longitude: 100.5018},
			to:       Location{Latitude: 13.7653, Longitude: 100.5018},
			expected: 1,
			delta:    0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance := tt.from.DistanceTo(tt.to)
			if math.Abs(distance-tt.expected) > tt.delta {
				t.Errorf("DistanceTo() = %v km, expected %v km (±%v)", distance, tt.expected, tt.delta)
			}
		})
	}
}
