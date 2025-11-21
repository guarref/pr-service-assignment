package web

import (
	"net/http"

	"github.com/guarref/pr-service-assignment/internal/domain"
	"github.com/guarref/pr-service-assignment/internal/resperrors"
	"github.com/guarref/pr-service-assignment/internal/service"
	"github.com/guarref/pr-service-assignment/internal/web/odomains"
	"github.com/labstack/echo/v4"
)

type TeamHandler struct {
	service *service.TeamService
}

func NewTeamHandler(s *service.TeamService) *TeamHandler {
	return &TeamHandler{service: s}
}

// POST /team/add
func (h *TeamHandler) PostTeamAdd(ctx echo.Context) error {
	var body odomains.PostTeamAddJSONRequestBody

	if err := ctx.Bind(&body); err != nil {
		resp := NewErrorResponse(odomains.NOTFOUND, "invalid request body")
		return ctx.JSON(http.StatusBadRequest, resp)
	}

	team := domain.Team{
		TeamName: body.TeamName,
		Members:  make([]domain.TeamMember, 0, len(body.Members)),
	}

	for _, m := range body.Members {
		team.Members = append(team.Members, domain.TeamMember{
			UserID:   m.UserId,
			UserName: m.Username,
			IsActive: m.IsActive,
		})
	}

	if err := h.service.CreateTeam(ctx.Request().Context(), &team); err != nil {
		return writeDomainError(ctx, err)
	}

	respTeam := toOAPITeam(&team)

	return ctx.JSON(http.StatusCreated, struct {
		Team odomains.Team `json:"team"`
	}{
		Team: respTeam,
	})
}

// GET /team/get
func (h *TeamHandler) GetTeamGet(ctx echo.Context, params odomains.GetTeamGetParams) error {
	teamName := params.TeamName
	if teamName == "" {
		return writeDomainError(ctx, resperrors.ErrBadRequest)
	}

	team, err := h.service.GetTeamByName(ctx.Request().Context(), teamName)
	if err != nil {
		return writeDomainError(ctx, err)
	}

	respTeam := toOAPITeam(team)
	return ctx.JSON(http.StatusOK, respTeam)
}