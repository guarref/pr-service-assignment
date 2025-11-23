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

func TestPostPullRequestCreate_Success(t *testing.T) {
	e, _, mockUserRepo, mockPRRepo, _ := SetupTestApp()

	prReq := omodels.PostPullRequestCreateJSONRequestBody{
		PullRequestId:   "pr-1",
		PullRequestName: "Add feature",
		AuthorId:        "author1",
	}

	author := CreateTestUser("author1", "Author", "backend", true)
	activeUsers := []*models.User{
		CreateTestUser("reviewer1", "Reviewer1", "backend", true),
		CreateTestUser("reviewer2", "Reviewer2", "backend", true),
		CreateTestUser("reviewer3", "Reviewer3", "backend", true),
	}

	expectedPR := CreateTestPullRequest("pr-1", "Add feature", "author1", models.PullRequestOpen, []string{"reviewer1", "reviewer2"})

	mockUserRepo.On("GetUserByID", GetTestContext(), "author1").Return(author, nil)
	mockUserRepo.On("GetActiveUsersByTeam", GetTestContext(), "backend", "author1").Return(activeUsers, nil)
	mockPRRepo.On("CreatePullRequest", GetTestContext(), mock.MatchedBy(func(pr *models.PullRequest) bool {
		return pr.PullRequestID == "pr-1" && pr.AuthorID == "author1"
	})).Return(expectedPR, nil)

	body, _ := json.Marshal(prReq)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewPullRequestHandler(service.NewPullRequestService(mockPRRepo, mockUserRepo))
	err := handler.PostPullRequestCreate(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response struct {
		PR omodels.PullRequest `json:"pr"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "pr-1", response.PR.PullRequestId)
	assert.Equal(t, "Add feature", response.PR.PullRequestName)
	assert.Equal(t, "author1", response.PR.AuthorId)
	assert.Equal(t, omodels.PullRequestStatusOPEN, response.PR.Status)
	assert.LessOrEqual(t, len(response.PR.AssignedReviewers), 2)

	mockUserRepo.AssertExpectations(t)
	mockPRRepo.AssertExpectations(t)
}

func TestPostPullRequestCreate_Duplicate(t *testing.T) {
	e, _, mockUserRepo, mockPRRepo, _ := SetupTestApp()

	prReq := omodels.PostPullRequestCreateJSONRequestBody{
		PullRequestId:   "pr-2",
		PullRequestName: "Fix bug",
		AuthorId:        "author2",
	}

	author := CreateTestUser("author2", "Author2", "frontend", true)
	activeUsers := []*models.User{
		CreateTestUser("reviewer4", "Reviewer4", "frontend", true),
	}

	mockUserRepo.On("GetUserByID", GetTestContext(), "author2").Return(author, nil)
	mockUserRepo.On("GetActiveUsersByTeam", GetTestContext(), "frontend", "author2").Return(activeUsers, nil)
	mockPRRepo.On("CreatePullRequest", GetTestContext(), mock.Anything).Return(nil, errs.ErrPullRequestExists)

	body, _ := json.Marshal(prReq)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewPullRequestHandler(service.NewPullRequestService(mockPRRepo, mockUserRepo))
	err := handler.PostPullRequestCreate(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var errorResp omodels.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, omodels.PREXISTS, errorResp.Error.Code)

	mockUserRepo.AssertExpectations(t)
	mockPRRepo.AssertExpectations(t)
}

func TestPostPullRequestCreate_AuthorNotFound(t *testing.T) {
	e, _, mockUserRepo, mockPRRepo, _ := SetupTestApp()

	prReq := omodels.PostPullRequestCreateJSONRequestBody{
		PullRequestId:   "pr-3",
		PullRequestName: "Test PR",
		AuthorId:        "nonexistent-author",
	}

	mockUserRepo.On("GetUserByID", GetTestContext(), "nonexistent-author").Return(nil, errs.ErrUserNotFound)

	body, _ := json.Marshal(prReq)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewPullRequestHandler(service.NewPullRequestService(mockPRRepo, mockUserRepo))
	err := handler.PostPullRequestCreate(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockUserRepo.AssertExpectations(t)
}

func TestPostPullRequestMerge_Success(t *testing.T) {
	e, _, _, mockPRRepo, _ := SetupTestApp()

	mergeReq := omodels.PostPullRequestMergeJSONRequestBody{
		PullRequestId: "pr-4",
	}

	expectedPR := CreateTestPullRequest("pr-4", "Merge test", "author3", models.PullRequestMerged, []string{"reviewer5"})

	mockPRRepo.On("MergePullRequestByID", GetTestContext(), "pr-4").Return(expectedPR, nil)

	body, _ := json.Marshal(mergeReq)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewPullRequestHandler(service.NewPullRequestService(mockPRRepo, nil))
	err := handler.PostPullRequestMerge(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		PR omodels.PullRequest `json:"pr"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, omodels.PullRequestStatusMERGED, response.PR.Status)
	assert.NotNil(t, response.PR.MergedAt)

	mockPRRepo.AssertExpectations(t)
}

func TestPostPullRequestMerge_NotFound(t *testing.T) {
	e, _, _, mockPRRepo, _ := SetupTestApp()

	mergeReq := omodels.PostPullRequestMergeJSONRequestBody{
		PullRequestId: "nonexistent-pr",
	}

	mockPRRepo.On("MergePullRequestByID", GetTestContext(), "nonexistent-pr").Return(nil, errs.ErrPullRequestNotFound)

	body, _ := json.Marshal(mergeReq)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewPullRequestHandler(service.NewPullRequestService(mockPRRepo, nil))
	err := handler.PostPullRequestMerge(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockPRRepo.AssertExpectations(t)
}

func TestPostPullRequestReassign_Success(t *testing.T) {
	e, _, _, mockPRRepo, _ := SetupTestApp()

	reassignReq := omodels.PostPullRequestReassignJSONRequestBody{
		PullRequestId: "pr-5",
		OldUserId:     "reviewer1",
	}

	expectedPR := CreateTestPullRequest("pr-5", "Reassign test", "author4", models.PullRequestOpen, []string{"reviewer2", "reviewer3"})
	newReviewerID := "reviewer3"

	mockPRRepo.On("ReassignToPullRequest", GetTestContext(), "pr-5", "reviewer1").Return(expectedPR, newReviewerID, nil)

	body, _ := json.Marshal(reassignReq)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewPullRequestHandler(service.NewPullRequestService(mockPRRepo, nil))
	err := handler.PostPullRequestReassign(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		PR         omodels.PullRequest `json:"pr"`
		ReplacedBy string              `json:"replaced_by"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, newReviewerID, response.ReplacedBy)
	assert.NotContains(t, response.PR.AssignedReviewers, "reviewer1")
	assert.Contains(t, response.PR.AssignedReviewers, newReviewerID)

	mockPRRepo.AssertExpectations(t)
}

func TestPostPullRequestReassign_MergedPR(t *testing.T) {
	e, _, _, mockPRRepo, _ := SetupTestApp()

	reassignReq := omodels.PostPullRequestReassignJSONRequestBody{
		PullRequestId: "pr-6",
		OldUserId:     "reviewer2",
	}

	mockPRRepo.On("ReassignToPullRequest", GetTestContext(), "pr-6", "reviewer2").Return(nil, "", errs.ErrPullRequestMerged)

	body, _ := json.Marshal(reassignReq)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewPullRequestHandler(service.NewPullRequestService(mockPRRepo, nil))
	err := handler.PostPullRequestReassign(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var errorResp omodels.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, omodels.PRMERGED, errorResp.Error.Code)

	mockPRRepo.AssertExpectations(t)
}

func TestPostPullRequestReassign_NotAssigned(t *testing.T) {
	e, _, _, mockPRRepo, _ := SetupTestApp()

	reassignReq := omodels.PostPullRequestReassignJSONRequestBody{
		PullRequestId: "pr-7",
		OldUserId:     "not-reviewer",
	}

	mockPRRepo.On("ReassignToPullRequest", GetTestContext(), "pr-7", "not-reviewer").Return(nil, "", errs.ErrNotAssigned)

	body, _ := json.Marshal(reassignReq)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewPullRequestHandler(service.NewPullRequestService(mockPRRepo, nil))
	err := handler.PostPullRequestReassign(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var errorResp omodels.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, omodels.NOTASSIGNED, errorResp.Error.Code)

	mockPRRepo.AssertExpectations(t)
}

func TestGetUsersGetReview_Success(t *testing.T) {
	e, _, _, mockPRRepo, _ := SetupTestApp()

	expectedPRs := []*models.PullRequestShort{
		CreateTestPullRequestShort("pr-8", "PR 1", "author1", models.PullRequestOpen),
		CreateTestPullRequestShort("pr-9", "PR 2", "author2", models.PullRequestOpen),
	}

	mockPRRepo.On("GetPullRequestByReviewerID", GetTestContext(), "reviewer1").Return(expectedPRs, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=reviewer1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewPullRequestHandler(service.NewPullRequestService(mockPRRepo, nil))
	err := handler.GetUsersGetReview(c, omodels.GetUsersGetReviewParams{UserId: "reviewer1"})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		UserId       string                      `json:"user_id"`
		PullRequests []omodels.PullRequestShort `json:"pull_requests"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "reviewer1", response.UserId)
	assert.Len(t, response.PullRequests, 2)

	mockPRRepo.AssertExpectations(t)
}

func TestGetUsersGetReview_Empty(t *testing.T) {
	e, _, _, mockPRRepo, _ := SetupTestApp()

	mockPRRepo.On("GetPullRequestByReviewerID", GetTestContext(), "reviewer2").Return([]*models.PullRequestShort{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=reviewer2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := web.NewPullRequestHandler(service.NewPullRequestService(mockPRRepo, nil))
	err := handler.GetUsersGetReview(c, omodels.GetUsersGetReviewParams{UserId: "reviewer2"})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		UserId       string                      `json:"user_id"`
		PullRequests []omodels.PullRequestShort `json:"pull_requests"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "reviewer2", response.UserId)
	assert.Len(t, response.PullRequests, 0)

	mockPRRepo.AssertExpectations(t)
}
