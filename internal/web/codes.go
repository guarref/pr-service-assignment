package web

import (
	"errors"
	"net/http"
	"time"

	"github.com/guarref/pr-service-assignment/internal/domain"
	"github.com/guarref/pr-service-assignment/internal/resperrors"
	"github.com/guarref/pr-service-assignment/internal/web/odomains"
	"github.com/labstack/echo/v4"
)

// NewErrorResponse – сборка ErrorResponse из кода и сообщения.
func NewErrorResponse(code odomains.ErrorResponseErrorCode, message string) odomains.ErrorResponse {
	
	var er odomains.ErrorResponse
	er.Error.Code = code
	er.Error.Message = message

	return er
}

// writeDomainError – единое место, где мы превращаем доменные ошибки в HTTP-ответы.
func writeDomainError(c echo.Context, err error) error {
	
	if err == nil {
		return nil
	}

	// Валидация входных данных – просто 400 с сообщением
	if errors.Is(err, resperrors.ErrBadRequest) {
		resp := NewErrorResponse(odomains.NOTFOUND, err.Error())
		return c.JSON(http.StatusBadRequest, resp)
	}

	// 409 / 400 — доменные конфликты
	switch {
	case errors.Is(err, resperrors.ErrTeamExists):
		resp := NewErrorResponse(odomains.TEAMEXISTS, "team_name already exists")
		// В openapi для TEAM_EXISTS стоит 400
		return c.JSON(http.StatusBadRequest, resp)

	case errors.Is(err, resperrors.ErrPullRequestExists):
		resp := NewErrorResponse(odomains.PREXISTS, "PR id already exists")
		return c.JSON(http.StatusConflict, resp)

	case errors.Is(err, resperrors.ErrPullRequestMerged):
		resp := NewErrorResponse(odomains.PRMERGED, "cannot reassign on merged PR")
		return c.JSON(http.StatusConflict, resp)

	case errors.Is(err, resperrors.ErrNotAssigned):
		resp := NewErrorResponse(odomains.NOTASSIGNED, "reviewer is not assigned to this PR")
		return c.JSON(http.StatusConflict, resp)

	case errors.Is(err, resperrors.ErrNoCandidate):
		resp := NewErrorResponse(odomains.NOCANDIDATE, "no active replacement candidate in team")
		return c.JSON(http.StatusConflict, resp)
	}

	// 404 – любой вид not found
	if errors.Is(err, resperrors.ErrUserNotFound) ||
		errors.Is(err, resperrors.ErrTeamNotFound) ||
		errors.Is(err, resperrors.ErrPullRequestNotFound) {
		resp := NewErrorResponse(odomains.NOTFOUND, "resource not found")
		return c.JSON(http.StatusNotFound, resp)
	}

	// Фолбэк – 500 с generic сообщением
	resp := NewErrorResponse(odomains.NOTFOUND, "internal server error")
	return c.JSON(http.StatusInternalServerError, resp)
}

// Маппинги домен → openapi

func toOAPITeam(t *domain.Team) odomains.Team {
	members := make([]odomains.TeamMember, 0, len(t.Members))
	for _, m := range t.Members {
		members = append(members, odomains.TeamMember{
			UserId:   m.UserID,
			Username: m.UserName,
			IsActive: m.IsActive,
		})
	}

	return odomains.Team{
		TeamName: t.TeamName,
		Members:  members,
	}
}

func toOAPIUser(u *domain.User) odomains.User {
	return odomains.User{
		UserId:   u.UserID,
		Username: u.UserName,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}
}

func toOAPIPullRequest(pr *domain.PullRequest) odomains.PullRequest {
	var createdAt *time.Time
	if pr.CreatedAt != nil {
		t := *pr.CreatedAt
		createdAt = &t
	}

// MergedAt: *sql.NullTime -> *time.Time
    var mergedAt *time.Time
    if pr.MergedAt != nil && pr.MergedAt.Valid {
        t := pr.MergedAt.Time
        mergedAt = &t
    }

	return odomains.PullRequest{
		PullRequestId:     pr.PullRequestID,
		PullRequestName:   pr.PullRequestName,
		AuthorId:          pr.AuthorID,
		Status:            odomains.PullRequestStatus(pr.Status),
		AssignedReviewers: append([]string(nil), pr.AssignedReviewers...),
		CreatedAt:         createdAt,
		MergedAt:          mergedAt,
	}
}

func toOAPIPullRequestShort(pr *domain.PullRequestShort) odomains.PullRequestShort {
	return odomains.PullRequestShort{
		PullRequestId:   pr.PullRequestID,
		PullRequestName: pr.PullRequestName,
		AuthorId:        pr.AuthorID,
		Status:          odomains.PullRequestShortStatus(pr.Status),
	}
}

func toOAPIPullRequestShortList(prs []*domain.PullRequestShort) []odomains.PullRequestShort {
	result := make([]odomains.PullRequestShort, 0, len(prs))
	for _, pr := range prs {
		result = append(result, toOAPIPullRequestShort(pr))
	}
	return result
}
