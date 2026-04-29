package mysql

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

// GetOrCreateStats получает статистику или создает новую.
func (c *Client) GetOrCreateStats(testName string, questionCount int64) (*entities.TestStats, error) {
	newStats := entities.NewTestStats(testName, questionCount)

	// Маршалим распределения в JSON.
	var percentDistrib json.RawMessage
	var timeDistrib json.RawMessage

	if newStats.PercentDistrib != nil {
		data, err := json.Marshal(newStats.PercentDistrib)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal percent distrib: %w", err)
    }
		percentDistrib = data
	}
	if newStats.TimeDistrib != nil {
		data, err := json.Marshal(newStats.TimeDistrib)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal time distrib: %w", err)
    }
		timeDistrib = data
	}

	dbStats := &dbTestStats{
		TestName:       testName,
		UpdatedAt:      newStats.UpdatedAt,
		Attempts:       newStats.Attempts,
		AvgPercentage:  newStats.AvgPercentage,
		AvgTimeSpent:   newStats.AvgTimeSpent,
		MinTimeSpent:   newStats.MinTimeSpent,
		MaxTimeSpent:   newStats.MaxTimeSpent,
		PercentDistrib: percentDistrib,
		TimeDistrib:    timeDistrib,
	}

	// FirstOrCreate ищет по TestName, если нет — создает.
	result := c.db.Where(&dbTestStats{TestName: testName}).FirstOrCreate(dbStats)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get or create stats: %w", result.Error)
	}

	return c.GetStats(testName)
}

// GetStats получает статистику по имени теста.
func (c *Client) GetStats(testName string) (*entities.TestStats, error) {
	var dbStats dbTestStats
	err := c.db.Where("test_name = ?", testName).First(&dbStats).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entities.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	stats := &entities.TestStats{
		TestName:      dbStats.TestName,
		UpdatedAt:     dbStats.UpdatedAt,
		Attempts:      dbStats.Attempts,
		AvgPercentage: dbStats.AvgPercentage,
		AvgTimeSpent:  dbStats.AvgTimeSpent,
		MinTimeSpent:  dbStats.MinTimeSpent,
		MaxTimeSpent:  dbStats.MaxTimeSpent,
	}

	// Анмаршалим JSON в структуры.
	if len(dbStats.PercentDistrib) > 0 {
		var percentDistrib entities.PercentDistribution
		if err := json.Unmarshal(dbStats.PercentDistrib, &percentDistrib); err == nil {
			stats.PercentDistrib = &percentDistrib
		}
	}
	if len(dbStats.TimeDistrib) > 0 {
		var timeDistrib entities.TimeDistribution
		if err := json.Unmarshal(dbStats.TimeDistrib, &timeDistrib); err == nil {
			stats.TimeDistrib = &timeDistrib
		}
	}

	return stats, nil
}

// SaveStats сохраняет статистику.
func (c *Client) SaveStats(stats *entities.TestStats) error {
	var percentDistrib json.RawMessage
	var timeDistrib json.RawMessage

	if stats.PercentDistrib != nil {
		data, err := json.Marshal(stats.PercentDistrib)
    if err != nil {
        return fmt.Errorf("failed to marshal time distrib: %w", err)
    }
		percentDistrib = data
	}
	if stats.TimeDistrib != nil {
		data, err := json.Marshal(stats.TimeDistrib)
    if err != nil {
        return fmt.Errorf("failed to marshal time distrib: %w", err)
    }
		timeDistrib = data
	}

	dbStats := &dbTestStats{
		TestName:       stats.TestName,
		UpdatedAt:      time.Now(),
		Attempts:       stats.Attempts,
		AvgPercentage:  stats.AvgPercentage,
		AvgTimeSpent:   stats.AvgTimeSpent,
		MinTimeSpent:   stats.MinTimeSpent,
		MaxTimeSpent:   stats.MaxTimeSpent,
		PercentDistrib: percentDistrib,
		TimeDistrib:    timeDistrib,
	}

	return c.db.Save(dbStats).Error
}