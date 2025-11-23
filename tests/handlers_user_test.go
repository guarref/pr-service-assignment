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

func TestUserHandler_PostUsersSetIsActive(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*mocks.MockUserRepository)
		expectedStatus int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful set user active",
			requestBody: omodels.PostUsersSetIsActiveJSONRequestBody{
				UserId:   "user-1",
				IsActive: true,
			},
			setupMocks: func(userRepo *mocks.MockUserRepository) {
				user := &models.User{
					UserID:   "user-1",
					UserName: "testuser",
					TeamName: "team-1",
					IsActive: true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				userRepo.On("SetFlagIsActive", mock.Anything, "user-1", true).Return(user, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response struct {
					User omodels.User `json:"user"`
				}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "user-1", response.User.UserId)
				assert.Equal(t, "testuser", response.User.Username)
				assert.True(t, response.User.IsActive)
			},
		},
		{
			name: "successful set user inactive",
			requestBody: omodels.PostUsersSetIsActiveJSONRequestBody{
				UserId:   "user-1",
				IsActive: false,
			},
			setupMocks: func(userRepo *mocks.MockUserRepository) {
				user := &models.User{
					UserID:   "user-1",
					UserName: "testuser",
					TeamName: "team-1",
					IsActive: false,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				userRepo.On("SetFlagIsActive", mock.Anything, "user-1", false).Return(user, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response struct {
					User omodels.User `json:"user"`
				}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response.User.IsActive)
			},
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"invalid": "data",
			},
			setupMocks: func(userRepo *mocks.MockUserRepository) {
				// No mocks needed
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty user ID",
			requestBody: omodels.PostUsersSetIsActiveJSONRequestBody{
				UserId:   "",
				IsActive: true,
			},
			setupMocks: func(userRepo *mocks.MockUserRepository) {
				// Service will return ErrBadRequest
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "user not found",
			requestBody: omodels.PostUsersSetIsActiveJSONRequestBody{
				UserId:   "user-999",
				IsActive: true,
			},
			setupMocks: func(userRepo *mocks.MockUserRepository) {
				userRepo.On("SetFlagIsActive", mock.Anything, "user-999", true).Return(nil, errs.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			userRepo := new(mocks.MockUserRepository)
			userService := service.NewUserService(userRepo)
			handler := web.NewUserHandler(userService)

			tt.setupMocks(userRepo)

			// Create request
			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(bodyBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute
			err := handler.PostUsersSetIsActive(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.validateResponse != nil {
				tt.validateResponse(t, rec)
			}

			userRepo.AssertExpectations(t)
		})
	}
}

