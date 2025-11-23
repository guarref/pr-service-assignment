package web

import (
	"errors"
	"net/http"
	"time"

	"github.com/guarref/pr-service-assignment/internal/errs"
	"github.com/guarref/pr-service-assignment/internal/models"
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

	var respErr *errs.RespError
	if errors.As(err, &respErr) {
		var code omodels.ErrorResponseErrorCode

		switch respErr {
		case errs.ErrTeamExists:
			code = omodels.TEAMEXISTS
		case errs.ErrPullRequestExists:
			code = omodels.PREXISTS
		case errs.ErrPullRequestMerged:
			code = omodels.PRMERGED
		case errs.ErrNotAssigned:
			code = omodels.NOTASSIGNED
		case errs.ErrNoCandidate:
			code = omodels.NOCANDIDATE

		case errs.ErrTeamNotFound,
			errs.ErrUserNotFound,
			errs.ErrPullRequestNotFound,
			errs.ErrNotFound:
			code = omodels.NOTFOUND

		case errs.ErrBadRequest,
			errs.ErrInvalidJSON:
			code = omodels.NOTFOUND // вместо BAD REQUEST печаетаем NOT FOUND
		}

		resp := NewErrorResponse(code, respErr.Message)
		return c.JSON(respErr.StatusCode, resp)
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
	return omodels.User{
		UserId:   u.UserID,
		Username: u.UserName,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}
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
