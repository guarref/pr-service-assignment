package web

import (
	"net/http"

	"github.com/guarref/pr-service-assignment/internal/domain"
	"github.com/guarref/pr-service-assignment/internal/resperrors"
	"github.com/guarref/pr-service-assignment/internal/service"
	"github.com/guarref/pr-service-assignment/internal/web/odomains"
	"github.com/labstack/echo/v4"
)

type PullRequestHandler struct {
	service *service.PullRequestService
}

func NewPullRequestHandler(s *service.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{service: s}
}

// POST /pullRequest/create
func (h *PullRequestHandler) PostPullRequestCreate(ctx echo.Context) error {
	var body odomains.PostPullRequestCreateJSONRequestBody

	if err := ctx.Bind(&body); err != nil {
		resp := NewErrorResponse(odomains.NOTFOUND, "invalid request body")
		return ctx.JSON(http.StatusBadRequest, resp)
	}

	if body.PullRequestId == "" || body.PullRequestName == "" || body.AuthorId == "" {
		return writeDomainError(ctx, resperrors.ErrBadRequest)
	}

	pr := domain.PullRequest{
		PullRequestID:   body.PullRequestId,
		PullRequestName: body.PullRequestName,
		AuthorID:        body.AuthorId,
	}

	created, err := h.service.CreatePullRequest(ctx.Request().Context(), &pr)
	if err != nil {
		return writeDomainError(ctx, err)
	}

	respPR := toOAPIPullRequest(created)

	return ctx.JSON(http.StatusCreated, struct {
		PR odomains.PullRequest `json:"pr"`
	}{
		PR: respPR,
	})
}

// POST /pullRequest/merge
func (h *PullRequestHandler) PostPullRequestMerge(ctx echo.Context) error {
	var body odomains.PostPullRequestMergeJSONRequestBody

	if err := ctx.Bind(&body); err != nil {
		resp := NewErrorResponse(odomains.NOTFOUND, "invalid request body")
		return ctx.JSON(http.StatusBadRequest, resp)
	}

	if body.PullRequestId == "" {
		return writeDomainError(ctx, resperrors.ErrBadRequest)
	}

	pr, err := h.service.MergePullRequest(ctx.Request().Context(), body.PullRequestId)
	if err != nil {
		return writeDomainError(ctx, err)
	}

	respPR := toOAPIPullRequest(pr)

	return ctx.JSON(http.StatusOK, struct {
		PR odomains.PullRequest `json:"pr"`
	}{
		PR: respPR,
	})
}

// POST /pullRequest/reassign
func (h *PullRequestHandler) PostPullRequestReassign(ctx echo.Context) error {
	var body odomains.PostPullRequestReassignJSONRequestBody

	if err := ctx.Bind(&body); err != nil {
		resp := NewErrorResponse(odomains.NOTFOUND, "invalid request body")
		return ctx.JSON(http.StatusBadRequest, resp)
	}

	if body.PullRequestId == "" || body.OldUserId == "" {
		return writeDomainError(ctx, resperrors.ErrBadRequest)
	}

	pr, replacedBy, err := h.service.ReassignToPullRequest(ctx.Request().Context(), body.PullRequestId, body.OldUserId)
	if err != nil {
		return writeDomainError(ctx, err)
	}

	respPR := toOAPIPullRequest(pr)

	return ctx.JSON(http.StatusOK, struct {
		PR         odomains.PullRequest `json:"pr"`
		ReplacedBy string               `json:"replaced_by"`
	}{
		PR:         respPR,
		ReplacedBy: replacedBy,
	})
}

// GET /users/getReview
func (h *PullRequestHandler) GetUsersGetReview(ctx echo.Context, params odomains.GetUsersGetReviewParams) error {
	userID := params.UserId
	if userID == "" {
		return writeDomainError(ctx, resperrors.ErrBadRequest)
	}

	prs, err := h.service.GetPullRequestsByReviewer(ctx.Request().Context(), userID)
	if err != nil {
		return writeDomainError(ctx, err)
	}

	respList := toOAPIPullRequestShortList(prs)

	return ctx.JSON(http.StatusOK, struct {
		UserID       string                     `json:"user_id"`
		PullRequests []odomains.PullRequestShort `json:"pull_requests"`
	}{
		UserID:       userID,
		PullRequests: respList,
	})
}