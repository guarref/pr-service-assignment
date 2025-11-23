package web

import (
	"net/http"

	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/guarref/pr-service-assignment/internal/service"
	"github.com/guarref/pr-service-assignment/internal/web/omodels"
	"github.com/labstack/echo/v4"
)

type TeamHandler struct {
	service *service.TeamService
}

func NewTeamHandler(s *service.TeamService) *TeamHandler {
	return &TeamHandler{service: s}
}

// /team/add post
func (h *TeamHandler) PostTeamAdd(ctx echo.Context) error {

	var body omodels.PostTeamAddJSONRequestBody

	if err := ctx.Bind(&body); err != nil {
		resp := NewErrorResponse(omodels.NOTFOUND, "invalid request body")
		return ctx.JSON(http.StatusBadRequest, resp)
	}

	team := models.Team{
		TeamName: body.TeamName,
		Members:  make([]models.TeamMember, 0, len(body.Members)),
	}

	for _, m := range body.Members {
		team.Members = append(team.Members, models.TeamMember{
			UserID:   m.UserId,
			UserName: m.Username,
			IsActive: m.IsActive,
		})
	}

	if err := h.service.CreateTeam(ctx.Request().Context(), &team); err != nil {
		return mapErrorToHTTPResponse(ctx, err)
	}

	respTeam := toOAPITeam(&team)

	return ctx.JSON(http.StatusCreated, struct {
		Team omodels.Team `json:"team"`
	}{
		Team: respTeam,
	})
}

// /team/get get
func (h *TeamHandler) GetTeamGet(ctx echo.Context, params omodels.GetTeamGetParams) error {

	teamName := params.TeamName

	team, err := h.service.GetTeamByName(ctx.Request().Context(), teamName)
	if err != nil {
		return mapErrorToHTTPResponse(ctx, err)
	}

	respTeam := toOAPITeam(team)

	return ctx.JSON(http.StatusOK, respTeam)
}
