package web

import (
	"github.com/guarref/pr-service-assignment/internal/service"
	"github.com/guarref/pr-service-assignment/internal/web/omodels"
	"github.com/labstack/echo/v4"
)

// Router – единый тип, реализующий omodels.ServerInterface.
// Внутри делегирует в TeamHandler / UserHandler / PullRequestHandler.
type Router struct {
	teamHandler *TeamHandler
	userHandler *UserHandler
	prHandler   *PullRequestHandler
}

func NewRouter(teamSvc *service.TeamService, userSvc *service.UserService, prSvc *service.PullRequestService) *Router {
	return &Router{
		teamHandler: NewTeamHandler(teamSvc),
		userHandler: NewUserHandler(userSvc),
		prHandler:   NewPullRequestHandler(prSvc),
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

func (s *Router) GetUsersGetReview(ctx echo.Context, params omodels.GetUsersGetReviewParams) error {
	return s.prHandler.GetUsersGetReview(ctx, params)
}

func (s *Router) PostUsersSetIsActive(ctx echo.Context) error {
	return s.userHandler.PostUsersSetIsActive(ctx)
}

// RegisterRoutes – хелпер, который ты можешь вызывать из main.go.
func RegisterRoutes(e *echo.Echo, teamSvc *service.TeamService, userSvc *service.UserService, prSvc *service.PullRequestService) {
	
	server := NewRouter(teamSvc, userSvc, prSvc)
	omodels.RegisterHandlers(e, server)
}