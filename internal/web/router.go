package web

import (
	"github.com/guarref/pr-service-assignment/internal/service"
	"github.com/guarref/pr-service-assignment/internal/web/omodels"
	"github.com/labstack/echo/v4"
)

type Router struct {
	teamHandler  *TeamHandler
	userHandler  *UserHandler
	prHandler    *PullRequestHandler
	statsHandler *StatsHandler
}

func NewRouter(teamSvc *service.TeamService, userSvc *service.UserService, prSvc *service.PullRequestService, statsSvc *service.StatsService) *Router {
	return &Router{
		teamHandler:  NewTeamHandler(teamSvc),
		userHandler:  NewUserHandler(userSvc),
		prHandler:    NewPullRequestHandler(prSvc),
		statsHandler: NewStatsHandler(statsSvc),
	}
}

func (r *Router) PostPullRequestCreate(ctx echo.Context) error {
	return r.prHandler.PostPullRequestCreate(ctx)
}

func (r *Router) PostPullRequestMerge(ctx echo.Context) error {
	return r.prHandler.PostPullRequestMerge(ctx)
}

func (r *Router) PostPullRequestReassign(ctx echo.Context) error {
	return r.prHandler.PostPullRequestReassign(ctx)
}

func (r *Router) PostTeamAdd(ctx echo.Context) error {
	return r.teamHandler.PostTeamAdd(ctx)
}

func (r *Router) GetTeamGet(ctx echo.Context, params omodels.GetTeamGetParams) error {
	return r.teamHandler.GetTeamGet(ctx, params)
}

func (r *Router) GetUsersGetReview(ctx echo.Context, params omodels.GetUsersGetReviewParams) error {
	return r.prHandler.GetUsersGetReview(ctx, params)
}

func (r *Router) PostUsersSetIsActive(ctx echo.Context) error {
	return r.userHandler.PostUsersSetIsActive(ctx)
}

func (s *Router) GetStats(ctx echo.Context, params omodels.GetStatsParams) error {
	return s.statsHandler.GetStats(ctx, params)
}

func RegisterRoutes(e *echo.Echo, teamSvc *service.TeamService, userSvc *service.UserService, prSvc *service.PullRequestService, statsSvc *service.StatsService) {

	server := NewRouter(teamSvc, userSvc, prSvc, statsSvc)
	omodels.RegisterHandlers(e, server)
}
