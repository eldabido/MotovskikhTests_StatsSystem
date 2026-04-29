package manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAttempt(t *testing.T) {
	m := &Manager{}

	tests := []struct {
		name          string
		percentage    float64
		timeSpent     int64
		questionCount int64
		expectValid   bool
	}{
		{
			name:          "valid attempt with 30 regions",
			percentage:    70,
			timeSpent:     180,
			questionCount: 30,
			expectValid:   true,
		},
		{
			name:          "too fast for 30 regions",
			percentage:    70,
			timeSpent:     30,
			questionCount: 30,
			expectValid:   false,
		},
		{
			name:          "too low percentage",
			percentage:    3,
			timeSpent:     180,
			questionCount: 30,
			expectValid:   false,
		},
		{
			name:          "small test higher threshold",
			percentage:    8,
			timeSpent:     60,
			questionCount: 8,
			expectValid:   false,
		},
		{
			name:          "valid small test",
			percentage:    12,
			timeSpent:     30,
			questionCount: 8,
			expectValid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := m.validateAttempt(tt.name, tt.percentage, tt.timeSpent, tt.questionCount)
			assert.Equal(t, tt.expectValid, valid)
		})
	}
}