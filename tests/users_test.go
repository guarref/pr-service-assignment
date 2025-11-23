package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/guarref/pr-service-assignment/internal/errs"
	"github.com/guarref/pr-service-assignment/internal/service"
	"github.com/guarref/pr-service-assignment/internal/web"
	"github.com/guarref/pr-service-assignment/internal/web/omodels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostUsersSetIsActive_Success(t *testing.T) {
	e, _, mockUserRepo, _, _ := SetupTestApp()

	setActiveReq := omodels.PostUsersSetIsActiveJSONRequestBody{
		UserId:   "u1",
		IsActive: false,
	}

	expectedUser := CreateTestUser("u1", "Alice", "backend", false)

	mockUserRepo.On("SetFlagIsActive", GetTestContext(), "u1", false).Return(expectedUser, nil)

	body, _ := json.Marshal(setActiveReq)
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewUserHandler(service.NewUserService(mockUserRepo))
	err := handler.PostUsersSetIsActive(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		User omodels.User `json:"user"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "u1", response.User.UserId)
	assert.False(t, response.User.IsActive)

	mockUserRepo.AssertExpectations(t)
}

func TestPostUsersSetIsActive_Activate(t *testing.T) {
	e, _, mockUserRepo, _, _ := SetupTestApp()

	setActiveReq := omodels.PostUsersSetIsActiveJSONRequestBody{
		UserId:   "u2",
		IsActive: true,
	}

	expectedUser := CreateTestUser("u2", "Bob", "frontend", true)

	mockUserRepo.On("SetFlagIsActive", GetTestContext(), "u2", true).Return(expectedUser, nil)

	body, _ := json.Marshal(setActiveReq)
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewUserHandler(service.NewUserService(mockUserRepo))
	err := handler.PostUsersSetIsActive(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		User omodels.User `json:"user"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "u2", response.User.UserId)
	assert.True(t, response.User.IsActive)

	mockUserRepo.AssertExpectations(t)
}

func TestPostUsersSetIsActive_NotFound(t *testing.T) {
	e, _, mockUserRepo, _, _ := SetupTestApp()

	setActiveReq := omodels.PostUsersSetIsActiveJSONRequestBody{
		UserId:   "nonexistent",
		IsActive: false,
	}

	mockUserRepo.On("SetFlagIsActive", GetTestContext(), "nonexistent", false).Return(nil, errs.ErrUserNotFound)

	body, _ := json.Marshal(setActiveReq)
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewUserHandler(service.NewUserService(mockUserRepo))
	err := handler.PostUsersSetIsActive(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var errorResp omodels.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, omodels.NOTFOUND, errorResp.Error.Code)

	mockUserRepo.AssertExpectations(t)
}

func TestPostUsersSetIsActive_InvalidRequest(t *testing.T) {
	e, _, mockUserRepo, _, _ := SetupTestApp()

	// Отправляем невалидный JSON
	body := []byte(`{"invalid": "json"`)
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewUserHandler(service.NewUserService(mockUserRepo))
	err := handler.PostUsersSetIsActive(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
