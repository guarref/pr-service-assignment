package web

import (
	"errors"
	"net/http"
	"time"

	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/guarref/pr-service-assignment/internal/resperrors"
	"github.com/guarref/pr-service-assignment/internal/web/omodels"
	"github.com/labstack/echo/v4"
)

func NewErrorResponse(code omodels.ErrorResponseErrorCode, message string) omodels.ErrorResponse {

	var er omodels.ErrorResponse
	er.Error.Code = code
	er.Error.Message = message

	return er
}

func mapErrorToHTTPResponse(c echo.Context, err error) error {

	if err == nil {
		return nil
	}

	if errors.Is(err, resperrors.ErrBadRequest) {
		resp := NewErrorResponse(omodels.NOTFOUND, err.Error())
		return c.JSON(http.StatusBadRequest, resp)
	}

	switch {
	case errors.Is(err, resperrors.ErrTeamExists):
		resp := NewErrorResponse(omodels.TEAMEXISTS, "team_name already exists")
		return c.JSON(http.StatusBadRequest, resp)

	case errors.Is(err, resperrors.ErrPullRequestExists):
		resp := NewErrorResponse(omodels.PREXISTS, "PR id already exists")
		return c.JSON(http.StatusConflict, resp)

	case errors.Is(err, resperrors.ErrPullRequestMerged):
		resp := NewErrorResponse(omodels.PRMERGED, "cannot reassign on merged PR")
		return c.JSON(http.StatusConflict, resp)

	case errors.Is(err, resperrors.ErrNotAssigned):
		resp := NewErrorResponse(omodels.NOTASSIGNED, "reviewer is not assigned to this PR")
		return c.JSON(http.StatusConflict, resp)

	case errors.Is(err, resperrors.ErrNoCandidate):
		resp := NewErrorResponse(omodels.NOCANDIDATE, "no active replacement candidate in team")
		return c.JSON(http.StatusConflict, resp)
	}

	if errors.Is(err, resperrors.ErrUserNotFound) ||
		errors.Is(err, resperrors.ErrTeamNotFound) ||
		errors.Is(err, resperrors.ErrPullRequestNotFound) {
		resp := NewErrorResponse(omodels.NOTFOUND, "resource not found")
		return c.JSON(http.StatusNotFound, resp)
	}

	resp := NewErrorResponse(omodels.NOTFOUND, "internal server error")
	return c.JSON(http.StatusInternalServerError, resp)
}

func toOAPITeam(t *models.Team) omodels.Team {

	members := make([]omodels.TeamMember, 0, len(t.Members))
	for _, m := range t.Members {
		members = append(members, omodels.TeamMember{
			UserId:   m.UserID,
			Username: m.UserName,
			IsActive: m.IsActive,
		})
	}

	return omodels.Team{TeamName: t.TeamName, Members: members}
}

func toOAPIUser(u *models.User) omodels.User {
	return omodels.User{UserId: u.UserID, Username: u.UserName, TeamName: u.TeamName, IsActive: u.IsActive}
}

func toOAPIPullRequest(pr *models.PullRequest) omodels.PullRequest {

	var mergedAt *time.Time
	if pr.MergedAt != nil {
		t := *pr.MergedAt
		mergedAt = &t
	}
	createdAt := pr.CreatedAt

	return omodels.PullRequest{
		PullRequestId:     pr.PullRequestID,
		PullRequestName:   pr.PullRequestName,
		AuthorId:          pr.AuthorID,
		Status:            omodels.PullRequestStatus(pr.Status),
		AssignedReviewers: append([]string(nil), pr.AssignedReviewers...),
		CreatedAt:         &createdAt,
		MergedAt:          mergedAt,
	}
}

func toOAPIPullRequestShort(pr *models.PullRequestShort) omodels.PullRequestShort {
	return omodels.PullRequestShort{
		PullRequestId:   pr.PullRequestID,
		PullRequestName: pr.PullRequestName,
		AuthorId:        pr.AuthorID,
		Status:          omodels.PullRequestShortStatus(pr.Status),
	}
}

func toOAPIPullRequestShortList(prs []*models.PullRequestShort) []omodels.PullRequestShort {

	result := make([]omodels.PullRequestShort, 0, len(prs))
	for _, pr := range prs {
		result = append(result, toOAPIPullRequestShort(pr))
	}
	return result
}
