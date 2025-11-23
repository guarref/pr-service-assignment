package web

import (
	"net/http"

	"github.com/guarref/pr-service-assignment/internal/service"
	"github.com/guarref/pr-service-assignment/internal/web/omodels"
	"github.com/labstack/echo/v4"
)

type StatsHandler struct {
	service *service.StatsService
}

func NewStatsHandler(s *service.StatsService) *StatsHandler {
	return &StatsHandler{service: s}
}

// /stats get
func (h *StatsHandler) GetStats(ctx echo.Context, params omodels.GetStatsParams) error {

	var top *int

	if params.Top != nil {
		v := int(*params.Top)
		top = &v
	}

	stats, err := h.service.GetStats(ctx.Request().Context(), top)
	if err != nil {
		return mapErrorToHTTPResponse(ctx, err)
	}

	topReviewers := make([]omodels.TopReviewer, 0, len(stats.TopReviewers))
	for _, r := range stats.TopReviewers {
		topReviewers = append(topReviewers, omodels.TopReviewer{
			UserId:      r.UserID,
			Username:    r.UserName,
			ReviewCount: r.ReviewCount,
		})
	}

	resp := omodels.Stats{
		TotalTeams:        stats.TotalTeams,
		TotalUsers:        stats.TotalUsers,
		ActiveUsers:       stats.ActiveUsers,
		TotalPullRequests: stats.TotalPullRequests,
		OpenPullRequests:  stats.OpenPullRequests,
		TopReviewers:      topReviewers,
	}

	return ctx.JSON(http.StatusOK, resp)
}
