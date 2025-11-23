package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/guarref/pr-service-assignment/internal/errs"
	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/guarref/pr-service-assignment/internal/service"
	"github.com/guarref/pr-service-assignment/internal/web"
	"github.com/guarref/pr-service-assignment/internal/web/omodels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPostTeamAdd_Success(t *testing.T) {
	e, mockTeamRepo, _, _, _ := SetupTestApp()

	teamReq := omodels.PostTeamAddJSONRequestBody{
		TeamName: "backend",
		Members: []omodels.TeamMember{
			{UserId: "u1", Username: "Alice", IsActive: true},
			{UserId: "u2", Username: "Bob", IsActive: true},
		},
	}

	expectedTeam := CreateTestTeam("backend", []models.TeamMember{
		CreateTestTeamMember("u1", "Alice", true),
		CreateTestTeamMember("u2", "Bob", true),
	})

	mockTeamRepo.On("CreateTeam", GetTestContext(), mock.MatchedBy(func(team *models.Team) bool {
		return team.TeamName == "backend" && len(team.Members) == 2
	})).Return(nil).Run(func(args mock.Arguments) {
		team := args.Get(1).(*models.Team)
		team.CreatedAt = expectedTeam.CreatedAt
		team.UpdatedAt = expectedTeam.UpdatedAt
	})

	body, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewTeamHandler(service.NewTeamService(mockTeamRepo))
	err := handler.PostTeamAdd(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response struct {
		Team omodels.Team `json:"team"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "backend", response.Team.TeamName)
	assert.Len(t, response.Team.Members, 2)

	mockTeamRepo.AssertExpectations(t)
}

func TestPostTeamAdd_Duplicate(t *testing.T) {
	e, mockTeamRepo, _, _, _ := SetupTestApp()

	teamReq := omodels.PostTeamAddJSONRequestBody{
		TeamName: "frontend",
		Members: []omodels.TeamMember{
			{UserId: "u3", Username: "Charlie", IsActive: true},
		},
	}

	mockTeamRepo.On("CreateTeam", GetTestContext(), mock.MatchedBy(func(team *models.Team) bool {
		return team.TeamName == "frontend"
	})).Return(errs.ErrTeamExists)

	body, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewTeamHandler(service.NewTeamService(mockTeamRepo))
	err := handler.PostTeamAdd(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var errorResp omodels.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, omodels.TEAMEXISTS, errorResp.Error.Code)

	mockTeamRepo.AssertExpectations(t)
}

func TestGetTeamGet_Success(t *testing.T) {
	e, mockTeamRepo, _, _, _ := SetupTestApp()

	expectedTeam := CreateTestTeam("devops", []models.TeamMember{
		CreateTestTeamMember("u4", "David", true),
		CreateTestTeamMember("u5", "Eve", false),
	})

	mockTeamRepo.On("GetTeamByName", GetTestContext(), "devops").Return(expectedTeam, nil)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=devops", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/team/get")
	c.SetParamNames("team_name")
	c.SetParamValues("devops")

	handler := web.NewTeamHandler(service.NewTeamService(mockTeamRepo))
	err := handler.GetTeamGet(c, omodels.GetTeamGetParams{TeamName: "devops"})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var team omodels.Team
	err = json.Unmarshal(rec.Body.Bytes(), &team)
	require.NoError(t, err)
	assert.Equal(t, "devops", team.TeamName)
	assert.Len(t, team.Members, 2)

	mockTeamRepo.AssertExpectations(t)
}

func TestGetTeamGet_NotFound(t *testing.T) {
	e, mockTeamRepo, _, _, _ := SetupTestApp()

	mockTeamRepo.On("GetTeamByName", GetTestContext(), "nonexistent").Return(nil, errs.ErrTeamNotFound)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewTeamHandler(service.NewTeamService(mockTeamRepo))
	err := handler.GetTeamGet(c, omodels.GetTeamGetParams{TeamName: "nonexistent"})

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var errorResp omodels.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, omodels.NOTFOUND, errorResp.Error.Code)

	mockTeamRepo.AssertExpectations(t)
}

func TestPostTeamDeactivate_Success(t *testing.T) {
	e, mockTeamRepo, _, _, _ := SetupTestApp()

	deactivateReq := omodels.PostTeamDeactivateJSONRequestBody{
		TeamName: "qa",
		UserIds:  &[]string{"u6", "u7"},
	}

	expectedDeactivated := []string{"u6", "u7"}

	mockTeamRepo.On("DeactivateUsersAndReassignPRs", GetTestContext(), "qa", []string{"u6", "u7"}).Return(expectedDeactivated, nil)

	body, _ := json.Marshal(deactivateReq)
	req := httptest.NewRequest(http.MethodPost, "/team/deactivate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewTeamHandler(service.NewTeamService(mockTeamRepo))
	err := handler.PostTeamDeactivate(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response omodels.DeactivateUsersResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.DeactivatedUsers, 2)
	assert.Contains(t, response.DeactivatedUsers, "u6")
	assert.Contains(t, response.DeactivatedUsers, "u7")

	mockTeamRepo.AssertExpectations(t)
}

func TestPostTeamDeactivate_AllUsers(t *testing.T) {
	e, mockTeamRepo, _, _, _ := SetupTestApp()

	deactivateReq := omodels.PostTeamDeactivateJSONRequestBody{
		TeamName: "qa",
		UserIds:  nil,
	}

	expectedDeactivated := []string{"u8", "u9", "u10"}

	mockTeamRepo.On("DeactivateUsersAndReassignPRs", GetTestContext(), "qa", []string(nil)).Return(expectedDeactivated, nil)

	body, _ := json.Marshal(deactivateReq)
	req := httptest.NewRequest(http.MethodPost, "/team/deactivate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewTeamHandler(service.NewTeamService(mockTeamRepo))
	err := handler.PostTeamDeactivate(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response omodels.DeactivateUsersResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.DeactivatedUsers, 3)

	mockTeamRepo.AssertExpectations(t)
}

func TestPostTeamDeactivate_NotFound(t *testing.T) {
	e, mockTeamRepo, _, _, _ := SetupTestApp()

	deactivateReq := omodels.PostTeamDeactivateJSONRequestBody{
		TeamName: "nonexistent",
		UserIds:  &[]string{"u1"},
	}

	mockTeamRepo.On("DeactivateUsersAndReassignPRs", GetTestContext(), "nonexistent", []string{"u1"}).Return(nil, errs.ErrTeamNotFound)

	body, _ := json.Marshal(deactivateReq)
	req := httptest.NewRequest(http.MethodPost, "/team/deactivate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewTeamHandler(service.NewTeamService(mockTeamRepo))
	err := handler.PostTeamDeactivate(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockTeamRepo.AssertExpectations(t)
}
