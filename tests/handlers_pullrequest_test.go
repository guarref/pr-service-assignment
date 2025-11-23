package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/guarref/pr-service-assignment/internal/errs"
	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/guarref/pr-service-assignment/internal/repository/mocks"
	"github.com/guarref/pr-service-assignment/internal/service"
	"github.com/guarref/pr-service-assignment/internal/web"
	"github.com/guarref/pr-service-assignment/internal/web/omodels"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPullRequestHandler_PostPullRequestCreate(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      interface{}
		setupMocks       func(*mocks.MockPullRequestRepository, *mocks.MockUserRepository)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful PR creation",
			requestBody: omodels.PostPullRequestCreateJSONRequestBody{
				PullRequestId:   "pr-123",
				PullRequestName: "Test PR",
				AuthorId:        "user-1",
			},
			setupMocks: func(prRepo *mocks.MockPullRequestRepository, userRepo *mocks.MockUserRepository) {
				author := &models.User{
					UserID:   "user-1",
					UserName: "testuser",
					TeamName: "team-1",
					IsActive: true,
				}
				activeUsers := []*models.User{
					{UserID: "user-2", UserName: "reviewer1", TeamName: "team-1", IsActive: true},
					{UserID: "user-3", UserName: "reviewer2", TeamName: "team-1", IsActive: true},
				}
				createdPR := &models.PullRequest{
					PullRequestID:     "pr-123",
					PullRequestName:   "Test PR",
					AuthorID:          "user-1",
					Status:            models.PullRequestOpen,
					AssignedReviewers: []string{"user-2", "user-3"},
					CreatedAt:         time.Now(),
				}

				userRepo.On("GetUserByID", mock.Anything, "user-1").Return(author, nil)
				userRepo.On("GetActiveUsersByTeam", mock.Anything, "team-1", "user-1").Return(activeUsers, nil)
				prRepo.On("CreatePullRequest", mock.Anything, mock.MatchedBy(func(pr *models.PullRequest) bool {
					return pr.PullRequestID == "pr-123" && pr.PullRequestName == "Test PR" && pr.AuthorID == "user-1"
				})).Return(createdPR, nil)
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response struct {
					PR omodels.PullRequest `json:"pr"`
				}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "pr-123", response.PR.PullRequestId)
				assert.Equal(t, "Test PR", response.PR.PullRequestName)
				assert.Equal(t, omodels.PullRequestStatus(models.PullRequestOpen), response.PR.Status)
				assert.Len(t, response.PR.AssignedReviewers, 2)
			},
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"invalid": "data",
			},
			setupMocks: func(prRepo *mocks.MockPullRequestRepository, userRepo *mocks.MockUserRepository) {
				// No mocks needed for invalid request
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing required fields",
			requestBody: omodels.PostPullRequestCreateJSONRequestBody{
				PullRequestId: "",
			},
			setupMocks: func(prRepo *mocks.MockPullRequestRepository, userRepo *mocks.MockUserRepository) {
				// Service will return ErrBadRequest
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "author not found",
			requestBody: omodels.PostPullRequestCreateJSONRequestBody{
				PullRequestId:   "pr-123",
				PullRequestName: "Test PR",
				AuthorId:        "user-1",
			},
			setupMocks: func(prRepo *mocks.MockPullRequestRepository, userRepo *mocks.MockUserRepository) {
				userRepo.On("GetUserByID", mock.Anything, "user-1").Return(nil, errs.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "no active reviewers available",
			requestBody: omodels.PostPullRequestCreateJSONRequestBody{
				PullRequestId:   "pr-123",
				PullRequestName: "Test PR",
				AuthorId:        "user-1",
			},
			setupMocks: func(prRepo *mocks.MockPullRequestRepository, userRepo *mocks.MockUserRepository) {
				author := &models.User{
					UserID:   "user-1",
					UserName: "testuser",
					TeamName: "team-1",
					IsActive: true,
				}
				activeUsers := []*models.User{} // No active users
				createdPR := &models.PullRequest{
					PullRequestID:     "pr-123",
					PullRequestName:   "Test PR",
					AuthorID:          "user-1",
					Status:            models.PullRequestOpen,
					AssignedReviewers: []string{},
					CreatedAt:         time.Now(),
				}

				userRepo.On("GetUserByID", mock.Anything, "user-1").Return(author, nil)
				userRepo.On("GetActiveUsersByTeam", mock.Anything, "team-1", "user-1").Return(activeUsers, nil)
				prRepo.On("CreatePullRequest", mock.Anything, mock.Anything).Return(createdPR, nil)
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response struct {
					PR omodels.PullRequest `json:"pr"`
				}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Empty(t, response.PR.AssignedReviewers)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			prRepo := new(mocks.MockPullRequestRepository)
			userRepo := new(mocks.MockUserRepository)
			prService := service.NewPullRequestService(prRepo, userRepo)
			handler := web.NewPullRequestHandler(prService)

			tt.setupMocks(prRepo, userRepo)

			// Create request
			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(bodyBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute
			err := handler.PostPullRequestCreate(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.validateResponse != nil {
				tt.validateResponse(t, rec)
			}

			prRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}

func TestPullRequestHandler_PostPullRequestMerge(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      interface{}
		setupMocks       func(*mocks.MockPullRequestRepository)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful PR merge",
			requestBody: omodels.PostPullRequestMergeJSONRequestBody{
				PullRequestId: "pr-123",
			},
			setupMocks: func(prRepo *mocks.MockPullRequestRepository) {
				mergedAt := time.Now()
				mergedPR := &models.PullRequest{
					PullRequestID:     "pr-123",
					PullRequestName:   "Test PR",
					AuthorID:          "user-1",
					Status:            models.PullRequestMerged,
					AssignedReviewers: []string{"user-2", "user-3"},
					CreatedAt:         time.Now().Add(-time.Hour),
					MergedAt:          &mergedAt,
				}
				prRepo.On("MergePullRequestByID", mock.Anything, "pr-123").Return(mergedPR, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response struct {
					PR omodels.PullRequest `json:"pr"`
				}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "pr-123", response.PR.PullRequestId)
				assert.Equal(t, omodels.PullRequestStatus(models.PullRequestMerged), response.PR.Status)
				assert.NotNil(t, response.PR.MergedAt)
			},
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"invalid": "data",
			},
			setupMocks: func(prRepo *mocks.MockPullRequestRepository) {
				// No mocks needed
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "PR not found",
			requestBody: omodels.PostPullRequestMergeJSONRequestBody{
				PullRequestId: "pr-999",
			},
			setupMocks: func(prRepo *mocks.MockPullRequestRepository) {
				prRepo.On("MergePullRequestByID", mock.Anything, "pr-999").Return(nil, errs.ErrPullRequestNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			prRepo := new(mocks.MockPullRequestRepository)
			userRepo := new(mocks.MockUserRepository)
			prService := service.NewPullRequestService(prRepo, userRepo)
			handler := web.NewPullRequestHandler(prService)

			tt.setupMocks(prRepo)

			// Create request
			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(bodyBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute
			err := handler.PostPullRequestMerge(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.validateResponse != nil {
				tt.validateResponse(t, rec)
			}

			prRepo.AssertExpectations(t)
		})
	}
}

func TestPullRequestHandler_PostPullRequestReassign(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      interface{}
		setupMocks       func(*mocks.MockPullRequestRepository)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful PR reassignment",
			requestBody: omodels.PostPullRequestReassignJSONRequestBody{
				PullRequestId: "pr-123",
				OldUserId:     "user-2",
			},
			setupMocks: func(prRepo *mocks.MockPullRequestRepository) {
				reassignedPR := &models.PullRequest{
					PullRequestID:     "pr-123",
					PullRequestName:   "Test PR",
					AuthorID:          "user-1",
					Status:            models.PullRequestOpen,
					AssignedReviewers: []string{"user-3", "user-4"},
					CreatedAt:         time.Now(),
				}
				prRepo.On("ReassignToPullRequest", mock.Anything, "pr-123", "user-2").Return(reassignedPR, "user-4", nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response struct {
					PR         omodels.PullRequest `json:"pr"`
					ReplacedBy string              `json:"replaced_by"`
				}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "pr-123", response.PR.PullRequestId)
				assert.Equal(t, "user-4", response.ReplacedBy)
			},
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"invalid": "data",
			},
			setupMocks: func(prRepo *mocks.MockPullRequestRepository) {
				// No mocks needed
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "PR not found",
			requestBody: omodels.PostPullRequestReassignJSONRequestBody{
				PullRequestId: "pr-999",
				OldUserId:     "user-2",
			},
			setupMocks: func(prRepo *mocks.MockPullRequestRepository) {
				prRepo.On("ReassignToPullRequest", mock.Anything, "pr-999", "user-2").Return(nil, "", errs.ErrPullRequestNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			prRepo := new(mocks.MockPullRequestRepository)
			userRepo := new(mocks.MockUserRepository)
			prService := service.NewPullRequestService(prRepo, userRepo)
			handler := web.NewPullRequestHandler(prService)

			tt.setupMocks(prRepo)

			// Create request
			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(bodyBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute
			err := handler.PostPullRequestReassign(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.validateResponse != nil {
				tt.validateResponse(t, rec)
			}

			prRepo.AssertExpectations(t)
		})
	}
}

func TestPullRequestHandler_GetUsersGetReview(t *testing.T) {
	tests := []struct {
		name             string
		userID           string
		setupMocks       func(*mocks.MockPullRequestRepository)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "successful get PRs by reviewer",
			userID: "user-2",
			setupMocks: func(prRepo *mocks.MockPullRequestRepository) {
				prs := []*models.PullRequestShort{
					{
						PullRequestID:   "pr-1",
						PullRequestName: "PR 1",
						AuthorID:        "user-1",
						Status:          models.PullRequestOpen,
					},
					{
						PullRequestID:   "pr-2",
						PullRequestName: "PR 2",
						AuthorID:        "user-3",
						Status:          models.PullRequestOpen,
					},
				}
				prRepo.On("GetPullRequestByReviewerID", mock.Anything, "user-2").Return(prs, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response struct {
					UserID       string                     `json:"user_id"`
					PullRequests []omodels.PullRequestShort `json:"pull_requests"`
				}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "user-2", response.UserID)
				assert.Len(t, response.PullRequests, 2)
			},
		},
		{
			name:   "empty list",
			userID: "user-2",
			setupMocks: func(prRepo *mocks.MockPullRequestRepository) {
				prRepo.On("GetPullRequestByReviewerID", mock.Anything, "user-2").Return([]*models.PullRequestShort{}, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response struct {
					UserID       string                     `json:"user_id"`
					PullRequests []omodels.PullRequestShort `json:"pull_requests"`
				}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Empty(t, response.PullRequests)
			},
		},
		{
			name:   "missing user ID",
			userID: "",
			setupMocks: func(prRepo *mocks.MockPullRequestRepository) {
				// Service will return ErrBadRequest
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			prRepo := new(mocks.MockPullRequestRepository)
			userRepo := new(mocks.MockUserRepository)
			prService := service.NewPullRequestService(prRepo, userRepo)
			handler := web.NewPullRequestHandler(prService)

			tt.setupMocks(prRepo)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/users/getReview", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/users/getReview")
			c.SetParamNames("userId")
			c.SetParamValues(tt.userID)

			params := omodels.GetUsersGetReviewParams{
				UserId: tt.userID,
			}

			// Execute
			err := handler.GetUsersGetReview(c, params)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.validateResponse != nil {
				tt.validateResponse(t, rec)
			}

			prRepo.AssertExpectations(t)
		})
	}
}
