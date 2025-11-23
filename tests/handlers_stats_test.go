package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/guarref/pr-service-assignment/internal/repository/mocks"
	"github.com/guarref/pr-service-assignment/internal/service"
	"github.com/guarref/pr-service-assignment/internal/web"
	"github.com/guarref/pr-service-assignment/internal/web/omodels"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStatsHandler_GetStats(t *testing.T) {
	tests := []struct {
		name           string
		top            *int
		setupMocks     func(*mocks.MockStatsRepository)
		expectedStatus int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful get stats with top parameter",
			top:  intPtr(5),
			setupMocks: func(statsRepo *mocks.MockStatsRepository) {
				stats := &models.Stats{
					TotalTeams:        10,
					TotalUsers:        50,
					ActiveUsers:       45,
					TotalPullRequests: 100,
					OpenPullRequests:  20,
					TopReviewers: []*models.TopReviewer{
						{UserID: "user-1", UserName: "reviewer1", ReviewCount: 15},
						{UserID: "user-2", UserName: "reviewer2", ReviewCount: 12},
						{UserID: "user-3", UserName: "reviewer3", ReviewCount: 10},
					},
				}
				statsRepo.On("GetStats", mock.Anything, 5).Return(stats, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var stats omodels.Stats
				err := json.Unmarshal(rec.Body.Bytes(), &stats)
				assert.NoError(t, err)
				assert.Equal(t, 10, stats.TotalTeams)
				assert.Equal(t, 50, stats.TotalUsers)
				assert.Equal(t, 45, stats.ActiveUsers)
				assert.Equal(t, 100, stats.TotalPullRequests)
				assert.Equal(t, 20, stats.OpenPullRequests)
				assert.Len(t, stats.TopReviewers, 3)
			},
		},
		{
			name: "successful get stats without top parameter",
			top:  nil,
			setupMocks: func(statsRepo *mocks.MockStatsRepository) {
				stats := &models.Stats{
					TotalTeams:        10,
					TotalUsers:        50,
					ActiveUsers:       45,
					TotalPullRequests: 100,
					OpenPullRequests:  20,
					TopReviewers:      []*models.TopReviewer{},
				}
				statsRepo.On("GetStats", mock.Anything, 0).Return(stats, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var stats omodels.Stats
				err := json.Unmarshal(rec.Body.Bytes(), &stats)
				assert.NoError(t, err)
				assert.Equal(t, 10, stats.TotalTeams)
			},
		},
		{
			name: "top parameter exceeds limit",
			top:  intPtr(20), // Limit is 10
			setupMocks: func(statsRepo *mocks.MockStatsRepository) {
				stats := &models.Stats{
					TotalTeams:        10,
					TotalUsers:        50,
					ActiveUsers:       45,
					TotalPullRequests: 100,
					OpenPullRequests:  20,
					TopReviewers: []*models.TopReviewer{
						{UserID: "user-1", UserName: "reviewer1", ReviewCount: 15},
					},
				}
				statsRepo.On("GetStats", mock.Anything, 10).Return(stats, nil) // Should be capped at 10
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var stats omodels.Stats
				err := json.Unmarshal(rec.Body.Bytes(), &stats)
				assert.NoError(t, err)
				assert.Equal(t, 10, stats.TotalTeams)
			},
		},
		{
			name: "invalid top parameter (zero)",
			top:  intPtr(0),
			setupMocks: func(statsRepo *mocks.MockStatsRepository) {
				// Service will return ErrBadRequest
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid top parameter (negative)",
			top:  intPtr(-1),
			setupMocks: func(statsRepo *mocks.MockStatsRepository) {
				// Service will return ErrBadRequest
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty stats",
			top:  intPtr(5),
			setupMocks: func(statsRepo *mocks.MockStatsRepository) {
				stats := &models.Stats{
					TotalTeams:        0,
					TotalUsers:        0,
					ActiveUsers:       0,
					TotalPullRequests: 0,
					OpenPullRequests:  0,
					TopReviewers:      []*models.TopReviewer{},
				}
				statsRepo.On("GetStats", mock.Anything, 5).Return(stats, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var stats omodels.Stats
				err := json.Unmarshal(rec.Body.Bytes(), &stats)
				assert.NoError(t, err)
				assert.Equal(t, 0, stats.TotalTeams)
				assert.Empty(t, stats.TopReviewers)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			statsRepo := new(mocks.MockStatsRepository)
			statsService := service.NewStatsService(statsRepo)
			handler := web.NewStatsHandler(statsService)

			tt.setupMocks(statsRepo)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/stats", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/stats")

			var params omodels.GetStatsParams
			if tt.top != nil {
				params.Top = tt.top
			}

			// Execute
			err := handler.GetStats(c, params)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.validateResponse != nil {
				tt.validateResponse(t, rec)
			}

			statsRepo.AssertExpectations(t)
		})
	}
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}

