package manager

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/generated/mocks"
)

func TestSubmitTestResult_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockdb(ctrl)
	mgr := New(mockDB)

	testName := "europe"
	percentage := 75.0
	timeSpent := 180
	questionCount := 30

	percentBuckets := make(map[int]uint64)
	percentBuckets[0] = 10
	percentBuckets[20] = 20
	percentBuckets[40] = 30
	percentBuckets[60] = 25
	percentBuckets[80] = 15

	testStats := &entities.TestStats{
		TestName:      testName,
		Attempts:      100,
		AvgPercentage: 65.0,
		AvgTimeSpent:  200,
		MinTimeSpent:  30,
		MaxTimeSpent:  400,
		PercentDistrib: &entities.PercentDistribution{
			Buckets: percentBuckets,
		},
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
	}

	mockDB.EXPECT().
		GetOrCreateStats(testName, int64(questionCount)).
		Return(testStats, nil)

	mockDB.EXPECT().
		SaveStats(gomock.Any()).
		Return(nil)

	result, err := mgr.SubmitTestResult(testName, percentage, int64(timeSpent), int64(questionCount))

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Submitted)

	analysis := result.Analysis
	assert.InDelta(t, 75.0, analysis.Percentage, 0.001)
	assert.Equal(t, int64(180), analysis.TimeSpent)
}

func TestSubmitTestResult_InvalidAttempt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockdb(ctrl)
	mgr := New(mockDB)

	testName := "europe"
	percentage := 3.0
	timeSpent := 30
	questionCount := 30

	percentBuckets := make(map[int]uint64)
	for i := 0; i <= 100; i += 5 {
		percentBuckets[i] = 0
	}

	testStats := &entities.TestStats{
		TestName:       testName,
		Attempts:       0,
		AvgPercentage:  0,
		AvgTimeSpent:   0,
		MinTimeSpent:   0,
		MaxTimeSpent:   0,
		PercentDistrib: &entities.PercentDistribution{Buckets: percentBuckets},
		TimeDistrib:    &entities.TimeDistribution{Buckets: []entities.TimeBucket{}},
	}

	mockDB.EXPECT().
		GetOrCreateStats(testName, int64(questionCount)).
		Return(testStats, nil)

	result, err := mgr.SubmitTestResult(testName, percentage, int64(timeSpent), int64(questionCount))

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Submitted)
}