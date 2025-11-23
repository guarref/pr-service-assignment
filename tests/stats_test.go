package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/guarref/pr-service-assignment/internal/service"
	"github.com/guarref/pr-service-assignment/internal/web"
	"github.com/guarref/pr-service-assignment/internal/web/omodels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetStats_Success(t *testing.T) {
	e, _, _, _, mockStatsRepo := SetupTestApp()

	expectedStats := CreateTestStats(3, 10, 7, 15, 8, []*models.TopReviewer{
		CreateTestTopReviewer("u1", "Ivan", 4),
		CreateTestTopReviewer("u2", "Vasiliy", 3),
		CreateTestTopReviewer("u3", "Petr", 2),
	})

	mockStatsRepo.On("GetStats", GetTestContext(), 3).Return(expectedStats, nil)

	req := httptest.NewRequest(http.MethodGet, "/stats?top=3", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewStatsHandler(service.NewStatsService(mockStatsRepo))
	err := handler.GetStats(c, omodels.GetStatsParams{Top: func() *int { v := 3; return &v }()})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var stats omodels.Stats
	err = json.Unmarshal(rec.Body.Bytes(), &stats)
	require.NoError(t, err)
	assert.Equal(t, 3, stats.TotalTeams)
	assert.Equal(t, 10, stats.TotalUsers)
	assert.Equal(t, 7, stats.ActiveUsers)
	assert.Equal(t, 15, stats.TotalPullRequests)
	assert.Equal(t, 8, stats.OpenPullRequests)
	assert.Len(t, stats.TopReviewers, 3)

	mockStatsRepo.AssertExpectations(t)
}

func TestGetStats_DefaultTop(t *testing.T) {
	e, _, _, _, mockStatsRepo := SetupTestApp()

	expectedStats := CreateTestStats(2, 5, 4, 10, 6, []*models.TopReviewer{
		CreateTestTopReviewer("u1", "Ivan", 5),
		CreateTestTopReviewer("u2", "Vasiliy", 3),
	})

	mockStatsRepo.On("GetStats", GetTestContext(), 0).Return(expectedStats, nil)

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewStatsHandler(service.NewStatsService(mockStatsRepo))
	err := handler.GetStats(c, omodels.GetStatsParams{Top: nil})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var stats omodels.Stats
	err = json.Unmarshal(rec.Body.Bytes(), &stats)
	require.NoError(t, err)
	assert.Equal(t, 2, stats.TotalTeams)
	assert.Equal(t, 5, stats.TotalUsers)
	assert.Len(t, stats.TopReviewers, 2)

	mockStatsRepo.AssertExpectations(t)
}

func TestGetStats_LargeTop(t *testing.T) {
	e, _, _, _, mockStatsRepo := SetupTestApp()

	expectedStats := CreateTestStats(1, 3, 2, 5, 3, []*models.TopReviewer{
		CreateTestTopReviewer("u1", "Ivan", 2),
	})

	// Сервис ограничивает top до 10
	mockStatsRepo.On("GetStats", GetTestContext(), 10).Return(expectedStats, nil)

	req := httptest.NewRequest(http.MethodGet, "/stats?top=100", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewStatsHandler(service.NewStatsService(mockStatsRepo))
	top := 100
	err := handler.GetStats(c, omodels.GetStatsParams{Top: &top})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockStatsRepo.AssertExpectations(t)
}

func TestGetStats_InvalidTop(t *testing.T) {
	e, _, _, _, mockStatsRepo := SetupTestApp()

	req := httptest.NewRequest(http.MethodGet, "/stats?top=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewStatsHandler(service.NewStatsService(mockStatsRepo))
	top := 0
	err := handler.GetStats(c, omodels.GetStatsParams{Top: &top})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	mockStatsRepo.AssertNotCalled(t, "GetStats", mock.Anything, mock.Anything)
}

func TestGetStats_NegativeTop(t *testing.T) {
	e, _, _, _, mockStatsRepo := SetupTestApp()

	req := httptest.NewRequest(http.MethodGet, "/stats?top=-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewStatsHandler(service.NewStatsService(mockStatsRepo))
	top := -1
	err := handler.GetStats(c, omodels.GetStatsParams{Top: &top})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	mockStatsRepo.AssertNotCalled(t, "GetStats", mock.Anything, mock.Anything)
}
