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

func TestTeamHandler_PostTeamAdd(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*mocks.MockTeamRepository)
		expectedStatus int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful team creation",
			requestBody: omodels.PostTeamAddJSONRequestBody{
				TeamName: "team-1",
				Members: []omodels.TeamMember{
					{UserId: "user-1", Username: "user1", IsActive: true},
					{UserId: "user-2", Username: "user2", IsActive: true},
				},
			},
			setupMocks: func(teamRepo *mocks.MockTeamRepository) {
				teamRepo.On("CreateTeam", mock.Anything, mock.MatchedBy(func(team *models.Team) bool {
					return team.TeamName == "team-1" && len(team.Members) == 2
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response struct {
					Team omodels.Team `json:"team"`
				}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "team-1", response.Team.TeamName)
				assert.Len(t, response.Team.Members, 2)
			},
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"invalid": "data",
			},
			setupMocks: func(teamRepo *mocks.MockTeamRepository) {
				// No mocks needed
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty team name",
			requestBody: omodels.PostTeamAddJSONRequestBody{
				TeamName: "",
				Members: []omodels.TeamMember{
					{UserId: "user-1", Username: "user1", IsActive: true},
				},
			},
			setupMocks: func(teamRepo *mocks.MockTeamRepository) {
				// Service will return ErrBadRequest
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty members",
			requestBody: omodels.PostTeamAddJSONRequestBody{
				TeamName: "team-1",
				Members:  []omodels.TeamMember{},
			},
			setupMocks: func(teamRepo *mocks.MockTeamRepository) {
				// Service will return ErrBadRequest
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "team already exists",
			requestBody: omodels.PostTeamAddJSONRequestBody{
				TeamName: "team-1",
				Members: []omodels.TeamMember{
					{UserId: "user-1", Username: "user1", IsActive: true},
				},
			},
			setupMocks: func(teamRepo *mocks.MockTeamRepository) {
				teamRepo.On("CreateTeam", mock.Anything, mock.Anything).Return(errs.ErrTeamExists)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "invalid team name format",
			requestBody: omodels.PostTeamAddJSONRequestBody{
				TeamName: "team@1", // Invalid character
				Members: []omodels.TeamMember{
					{UserId: "user-1", Username: "user1", IsActive: true},
				},
			},
			setupMocks: func(teamRepo *mocks.MockTeamRepository) {
				// Service will return ErrBadRequest
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			teamRepo := new(mocks.MockTeamRepository)
			teamService := service.NewTeamService(teamRepo)
			handler := web.NewTeamHandler(teamService)

			tt.setupMocks(teamRepo)

			// Create request
			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(bodyBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute
			err := handler.PostTeamAdd(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.validateResponse != nil {
				tt.validateResponse(t, rec)
			}

			teamRepo.AssertExpectations(t)
		})
	}
}

func TestTeamHandler_GetTeamGet(t *testing.T) {
	tests := []struct {
		name           string
		teamName       string
		setupMocks     func(*mocks.MockTeamRepository)
		expectedStatus int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:     "successful get team",
			teamName: "team-1",
			setupMocks: func(teamRepo *mocks.MockTeamRepository) {
				team := &models.Team{
					TeamName: "team-1",
					Members: []models.TeamMember{
						{UserID: "user-1", UserName: "user1", IsActive: true},
						{UserID: "user-2", UserName: "user2", IsActive: true},
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				teamRepo.On("GetTeamByName", mock.Anything, "team-1").Return(team, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var team omodels.Team
				err := json.Unmarshal(rec.Body.Bytes(), &team)
				assert.NoError(t, err)
				assert.Equal(t, "team-1", team.TeamName)
				assert.Len(t, team.Members, 2)
			},
		},
		{
			name:     "team not found",
			teamName: "team-999",
			setupMocks: func(teamRepo *mocks.MockTeamRepository) {
				teamRepo.On("GetTeamByName", mock.Anything, "team-999").Return(nil, errs.ErrTeamNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "invalid team name",
			teamName: "team@1",
			setupMocks: func(teamRepo *mocks.MockTeamRepository) {
				// Service will return ErrBadRequest
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "empty team name",
			teamName: "",
			setupMocks: func(teamRepo *mocks.MockTeamRepository) {
				// Service will return ErrBadRequest
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			teamRepo := new(mocks.MockTeamRepository)
			teamService := service.NewTeamService(teamRepo)
			handler := web.NewTeamHandler(teamService)

			tt.setupMocks(teamRepo)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/team/get", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/team/get")
			c.SetParamNames("teamName")
			c.SetParamValues(tt.teamName)

			params := omodels.GetTeamGetParams{
				TeamName: tt.teamName,
			}

			// Execute
			err := handler.GetTeamGet(c, params)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.validateResponse != nil {
				tt.validateResponse(t, rec)
			}

			teamRepo.AssertExpectations(t)
		})
	}
}

func TestTeamHandler_PostTeamDeactivate(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*mocks.MockTeamRepository)
		expectedStatus int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful deactivation",
			requestBody: omodels.PostTeamDeactivateJSONRequestBody{
				TeamName: "team-1",
				UserIds: &[]string{"user-1", "user-2"},
			},
			setupMocks: func(teamRepo *mocks.MockTeamRepository) {
				deactivated := []string{"user-1", "user-2"}
				teamRepo.On("DeactivateUsersAndReassignPRs", mock.Anything, "team-1", []string{"user-1", "user-2"}).Return(deactivated, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response omodels.DeactivateUsersResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response.DeactivatedUsers, 2)
				assert.Contains(t, response.DeactivatedUsers, "user-1")
				assert.Contains(t, response.DeactivatedUsers, "user-2")
			},
		},
		{
			name: "deactivate all users (nil userIDs)",
			requestBody: omodels.PostTeamDeactivateJSONRequestBody{
				TeamName: "team-1",
				UserIds:  nil,
			},
			setupMocks: func(teamRepo *mocks.MockTeamRepository) {
				deactivated := []string{"user-1", "user-2", "user-3"}
				teamRepo.On("DeactivateUsersAndReassignPRs", mock.Anything, "team-1", []string(nil)).Return(deactivated, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response omodels.DeactivateUsersResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response.DeactivatedUsers, 3)
			},
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"invalid": "data",
			},
			setupMocks: func(teamRepo *mocks.MockTeamRepository) {
				// No mocks needed
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid team name",
			requestBody: omodels.PostTeamDeactivateJSONRequestBody{
				TeamName: "team@1",
				UserIds:  &[]string{"user-1"},
			},
			setupMocks: func(teamRepo *mocks.MockTeamRepository) {
				// Service will return ErrBadRequest
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "team not found",
			requestBody: omodels.PostTeamDeactivateJSONRequestBody{
				TeamName: "team-999",
				UserIds:  &[]string{"user-1"},
			},
			setupMocks: func(teamRepo *mocks.MockTeamRepository) {
				teamRepo.On("DeactivateUsersAndReassignPRs", mock.Anything, "team-999", []string{"user-1"}).Return(nil, errs.ErrTeamNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			teamRepo := new(mocks.MockTeamRepository)
			teamService := service.NewTeamService(teamRepo)
			handler := web.NewTeamHandler(teamService)

			tt.setupMocks(teamRepo)

			// Create request
			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/team/deactivate", bytes.NewReader(bodyBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute
			err := handler.PostTeamDeactivate(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.validateResponse != nil {
				tt.validateResponse(t, rec)
			}

			teamRepo.AssertExpectations(t)
		})
	}
}

