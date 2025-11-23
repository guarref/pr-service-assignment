package web

import (
	"net/http"

	"github.com/guarref/pr-service-assignment/internal/errs"
	"github.com/guarref/pr-service-assignment/internal/service"
	"github.com/guarref/pr-service-assignment/internal/web/omodels"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{service: s}
}

// /users/setIsActive post
func (h *UserHandler) PostUsersSetIsActive(ctx echo.Context) error {

	var body omodels.PostUsersSetIsActiveJSONRequestBody

	if err := ctx.Bind(&body); err != nil {
		return mapErrorToHTTPResponse(ctx, errs.ErrBadRequest)
	}

	user, err := h.service.SetFlagIsActive(ctx.Request().Context(), body.UserId, body.IsActive)
	if err != nil {
		return mapErrorToHTTPResponse(ctx, err)
	}

	respUser := toOAPIUser(user)

	return ctx.JSON(http.StatusOK, struct {
		User omodels.User `json:"user"`
	}{User: respUser})
}
