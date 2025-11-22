package web

import (
	"net/http"

	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/guarref/pr-service-assignment/internal/service"
	"github.com/guarref/pr-service-assignment/internal/web/omodels"
	"github.com/labstack/echo/v4"
)

type PullRequestHandler struct {
	service *service.PullRequestService
}

func NewPullRequestHandler(s *service.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{service: s}
}

// /pullRequest/create post
func (h *PullRequestHandler) PostPullRequestCreate(ctx echo.Context) error {

	var body omodels.PostPullRequestCreateJSONRequestBody

	if err := ctx.Bind(&body); err != nil {
		resp := NewErrorResponse(omodels.NOTFOUND, "invalid request body")
		return ctx.JSON(http.StatusBadRequest, resp)
	}

	pr := models.PullRequest{
		PullRequestID:   body.PullRequestId,
		PullRequestName: body.PullRequestName,
		AuthorID:        body.AuthorId,
	}

	created, err := h.service.CreatePullRequest(ctx.Request().Context(), &pr)
	if err != nil {
		return mapErrorToHTTPResponse(ctx, err)
	}

	respPR := toOAPIPullRequest(created)

	return ctx.JSON(http.StatusCreated, struct {
		PR omodels.PullRequest `json:"pr"`
	}{PR: respPR})
}

// /pullRequest/merge post
func (h *PullRequestHandler) PostPullRequestMerge(ctx echo.Context) error {

	var body omodels.PostPullRequestMergeJSONRequestBody

	if err := ctx.Bind(&body); err != nil {
		resp := NewErrorResponse(omodels.NOTFOUND, "invalid request body")
		return ctx.JSON(http.StatusBadRequest, resp)
	}

	pr, err := h.service.MergePullRequest(ctx.Request().Context(), body.PullRequestId)
	if err != nil {
		return mapErrorToHTTPResponse(ctx, err)
	}

	respPR := toOAPIPullRequest(pr)

	return ctx.JSON(http.StatusOK, struct {
		PR omodels.PullRequest `json:"pr"`
	}{PR: respPR})
}

// /pullRequest/reassign post
func (h *PullRequestHandler) PostPullRequestReassign(ctx echo.Context) error {

	var body omodels.PostPullRequestReassignJSONRequestBody

	if err := ctx.Bind(&body); err != nil {
		resp := NewErrorResponse(omodels.NOTFOUND, "invalid request body")
		return ctx.JSON(http.StatusBadRequest, resp)
	}

	pr, replacedBy, err := h.service.ReassignToPullRequest(ctx.Request().Context(), body.PullRequestId, body.OldUserId)
	if err != nil {
		return mapErrorToHTTPResponse(ctx, err)
	}

	respPR := toOAPIPullRequest(pr)

	return ctx.JSON(http.StatusOK, struct {
		PR         omodels.PullRequest `json:"pr"`
		ReplacedBy string              `json:"replaced_by"`
	}{
		PR:         respPR,
		ReplacedBy: replacedBy,
	})
}

// /users/getReview get
func (h *PullRequestHandler) GetUsersGetReview(ctx echo.Context, params omodels.GetUsersGetReviewParams) error {

	userID := params.UserId

	prs, err := h.service.GetPullRequestsByReviewer(ctx.Request().Context(), userID)
	if err != nil {
		return mapErrorToHTTPResponse(ctx, err)
	}

	respList := toOAPIPullRequestShortList(prs)

	return ctx.JSON(http.StatusOK, struct {
		UserID       string                     `json:"user_id"`
		PullRequests []omodels.PullRequestShort `json:"pull_requests"`
	}{
		UserID:       userID,
		PullRequests: respList,
	})
}
