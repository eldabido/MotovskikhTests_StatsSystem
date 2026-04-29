package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

const testEuropeID = "europe"

func setupTestDB(t *testing.T) *Client {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&dbTestStats{})
	require.NoError(t, err)

	return &Client{db: db}
}

func cleanupTestStats(t *testing.T, client *Client, testName string) {
	t.Helper()
	err := client.db.Exec("DELETE FROM db_test_stats WHERE test_name = ?", testName).Error
	require.NoError(t, err)
}

func TestGetOrCreateStats_CreateNew(t *testing.T) {
	client := setupTestDB(t)
	testName := testEuropeID
	defer cleanupTestStats(t, client, testName)

	stats, err := client.GetOrCreateStats(testName, 15)
	require.NoError(t, err)

	assert.NotNil(t, stats)
	assert.Equal(t, testName, stats.TestName)
	assert.Equal(t, uint64(0), stats.Attempts)
	assert.NotNil(t, stats.PercentDistrib)
	assert.NotNil(t, stats.TimeDistrib)

	// Должно быть 20 бакетов (5, 10, ..., 100)
	assert.Len(t, stats.PercentDistrib.Buckets, 20)
}

func TestGetOrCreateStats_GetExisting(t *testing.T) {
	client := setupTestDB(t)
	testName := testEuropeID
	defer cleanupTestStats(t, client, testName)

	// Создаем первую статистику
	stats1, err := client.GetOrCreateStats(testName, 15)
	require.NoError(t, err)

	// Обновляем данные
	stats1.Attempts = 5
	err = client.SaveStats(stats1)
	require.NoError(t, err)

	// Получаем существующую статистику
	stats2, err := client.GetOrCreateStats(testName, 15)
	require.NoError(t, err)

	assert.Equal(t, uint64(5), stats2.Attempts)
	assert.Equal(t, testName, stats2.TestName)
}

func TestGetStats_NotFound(t *testing.T) {
	client := setupTestDB(t)
	testName := "nonexistent"

	stats, err := client.GetStats(testName)
	require.Error(t, err)
	assert.Equal(t, entities.ErrNotFound, err)
	assert.Nil(t, stats)
}

func TestSaveStats_Update(t *testing.T) {
	client := setupTestDB(t)
	testName := testEuropeID
	defer cleanupTestStats(t, client, testName)

	// Создаем статистику
	stats, err := client.GetOrCreateStats(testName, 15)
	require.NoError(t, err)

	// Обновляем значения
	stats.Attempts = 10
	stats.AvgPercentage = 75.5
	stats.AvgTimeSpent = 180.0

	// Сохраняем
	err = client.SaveStats(stats)
	require.NoError(t, err)

	// Получаем и проверяем
	saved, err := client.GetStats(testName)
	require.NoError(t, err)

	assert.Equal(t, uint64(10), saved.Attempts)
	assert.InDelta(t, 75.5, saved.AvgPercentage, 0.001)
	assert.InDelta(t, 180.0, saved.AvgTimeSpent, 0.001)
}

func TestStatsWithDistributions(t *testing.T) {
	client := setupTestDB(t)
	testName := "test-dist"
	defer cleanupTestStats(t, client, testName)

	stats, err := client.GetOrCreateStats(testName, 10)
	require.NoError(t, err)

	// Проверяем процентные бакеты
	assert.NotNil(t, stats.PercentDistrib)
	assert.Len(t, stats.PercentDistrib.Buckets, 20)

	// Проверяем наличие ключей 5, 10, ..., 100
	for i := 5; i <= 100; i += 5 {
		_, ok := stats.PercentDistrib.Buckets[i]
		assert.True(t, ok, "missing key %d", i)
	}

	// Проверяем временные бакеты
	assert.NotNil(t, stats.TimeDistrib)
	assert.NotEmpty(t, stats.TimeDistrib.Buckets)
}

func TestSaveAndRetrieveStats(t *testing.T) {
	client := setupTestDB(t)
	testName := "test-save"
	defer cleanupTestStats(t, client, testName)

	// Создаем статистику
	stats := entities.NewTestStats(testName, 15)
	stats.Attempts = 100
	stats.AvgPercentage = 68.5
	stats.AvgTimeSpent = 150.5
	stats.MinTimeSpent = 60
	stats.MaxTimeSpent = 300

	// Сохраняем
	err := client.SaveStats(stats)
	require.NoError(t, err)

	// Получаем и проверяем
	retrieved, err := client.GetStats(testName)
	require.NoError(t, err)

	assert.Equal(t, stats.TestName, retrieved.TestName)
	assert.Equal(t, stats.Attempts, retrieved.Attempts)
	assert.InDelta(t, stats.AvgPercentage, retrieved.AvgPercentage, 0.001)
	assert.InDelta(t, stats.AvgTimeSpent, retrieved.AvgTimeSpent, 0.001)
	assert.Equal(t, stats.MinTimeSpent, retrieved.MinTimeSpent)
	assert.Equal(t, stats.MaxTimeSpent, retrieved.MaxTimeSpent)
}

func TestMultipleDifferentStats(t *testing.T) {
	client := setupTestDB(t)

	testNames := []string{"test1", "test2", "test3"}

	// Создаем статистику
	for _, id := range testNames {
		stats, err := client.GetOrCreateStats(id, 20)
		require.NoError(t, err)
		assert.NotNil(t, stats)
	}

	// Проверяем, что все создались
	for _, id := range testNames {
		stats, err := client.GetStats(id)
		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, id, stats.TestName)

		// Очищаем
		cleanupTestStats(t, client, id)
	}
}

func TestStatsAttemptsIncrement(t *testing.T) {
	client := setupTestDB(t)
	testName := "test-inc"
	defer cleanupTestStats(t, client, testName)

	// Создаем статистику
	stats, err := client.GetOrCreateStats(testName, 10)
	require.NoError(t, err)

	initialAttempts := stats.Attempts

	// Увеличиваем попытки
	stats.Attempts++
	err = client.SaveStats(stats)
	require.NoError(t, err)

	// Получаем и проверяем
	updated, err := client.GetStats(testName)
	require.NoError(t, err)
	assert.Equal(t, initialAttempts+1, updated.Attempts)
}