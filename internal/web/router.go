package web

import (
	"github.com/guarref/pr-service-assignment/internal/service"
	"github.com/guarref/pr-service-assignment/internal/web/odomains"
	"github.com/labstack/echo/v4"
)

// Router – единый тип, реализующий odomains.ServerInterface.
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

func (s *Router) PostPullRequestCreate(ctx echo.Context) error {
	return s.prHandler.PostPullRequestCreate(ctx)
}

func (s *Router) PostPullRequestMerge(ctx echo.Context) error {
	return s.prHandler.PostPullRequestMerge(ctx)
}

func (s *Router) PostPullRequestReassign(ctx echo.Context) error {
	return s.prHandler.PostPullRequestReassign(ctx)
}

func (s *Router) PostTeamAdd(ctx echo.Context) error {
	return s.teamHandler.PostTeamAdd(ctx)
}

func (s *Router) GetTeamGet(ctx echo.Context, params odomains.GetTeamGetParams) error {
	return s.teamHandler.GetTeamGet(ctx, params)
}

func (s *Router) GetUsersGetReview(ctx echo.Context, params odomains.GetUsersGetReviewParams) error {
	return s.prHandler.GetUsersGetReview(ctx, params)
}

func (s *Router) PostUsersSetIsActive(ctx echo.Context) error {
	return s.userHandler.PostUsersSetIsActive(ctx)
}

// RegisterRoutes – хелпер, который ты можешь вызывать из main.go.
func RegisterRoutes(e *echo.Echo, teamSvc *service.TeamService, userSvc *service.UserService, prSvc *service.PullRequestService) {
	
	server := NewRouter(teamSvc, userSvc, prSvc)
	odomains.RegisterHandlers(e, server)
}