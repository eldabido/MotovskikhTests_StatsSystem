package manager

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/generated/mocks"
)

func TestCalculatePercentile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockdb(ctrl)
	m := New(mockDB)

	buckets := make(map[int]uint64)
	for i := 0; i <= 100; i += 5 {
		if i < 50 {
			buckets[i] = 10
		} else {
			buckets[i] = 0
		}
	}
	buckets[0] = 10
	buckets[5] = 10
	buckets[10] = 10
	buckets[15] = 10
	buckets[20] = 10
	buckets[25] = 10
	buckets[30] = 10
	buckets[35] = 10
	buckets[40] = 10
	buckets[45] = 10

	tests := []struct {
		name       string
		stats      *entities.TestStats
		percentage float64
		expected   float64
	}{
		{
			name:       "nil stats returns 100",
			stats:      nil,
			percentage: 70,
			expected:   100,
		},
		{
			name: "empty stats returns 100",
			stats: &entities.TestStats{
				Attempts: 0,
				PercentDistrib: &entities.PercentDistribution{
					Buckets: buckets,
				},
			},
			percentage: 70,
			expected:   100,
		},
		{
			name: "average score calculates correctly",
			stats: &entities.TestStats{
				Attempts: 100,
				PercentDistrib: &entities.PercentDistribution{
					Buckets: buckets,
				},
			},
			percentage: 48,
			expected:   96,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.calculatePercentile(tt.stats, tt.percentage)
			assert.InDelta(t, tt.expected, result, 0.1)
		})
	}
}

func TestCalculateTimePercentile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockdb(ctrl)
	m := New(mockDB)

	tests := []struct {
		name      string
		stats     *entities.TestStats
		timeSpent int64
		expected  float64
	}{
		{
			name: "fast time returns low percentile",
			stats: &entities.TestStats{
				Attempts: 100,
				TimeDistrib: &entities.TimeDistribution{
					Buckets: []entities.TimeBucket{
						{MinSeconds: 0, Count: 30},
						{MinSeconds: 60, Count: 30},
						{MinSeconds: 120, Count: 20},
						{MinSeconds: 180, Count: 10},
						{MinSeconds: 240, Count: 5},
						{MinSeconds: 300, Count: 3},
						{MinSeconds: 360, Count: 2},
					},
				},
			},
			timeSpent: 45,
			expected:  77.5,
		},
		{
			name: "slow time returns high percentile",
			stats: &entities.TestStats{
				Attempts: 100,
				TimeDistrib: &entities.TimeDistribution{
					Buckets: []entities.TimeBucket{
						{MinSeconds: 0, Count: 30},
						{MinSeconds: 60, Count: 30},
						{MinSeconds: 120, Count: 20},
						{MinSeconds: 180, Count: 10},
						{MinSeconds: 240, Count: 5},
						{MinSeconds: 300, Count: 3},
						{MinSeconds: 360, Count: 2},
					},
				},
			},
			timeSpent: 400,
			expected:  0,
		},
		{
			name:      "nil stats returns 100",
			stats:     nil,
			timeSpent: 45,
			expected:  100,
		},
		{
			name: "empty stats returns 100",
			stats: &entities.TestStats{
				Attempts:    0,
				TimeDistrib: &entities.TimeDistribution{Buckets: []entities.TimeBucket{}},
			},
			timeSpent: 45,
			expected:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.calculateTimePercentile(tt.stats, tt.timeSpent)
			assert.InDelta(t, tt.expected, result, 0.1)
		})
	}
}