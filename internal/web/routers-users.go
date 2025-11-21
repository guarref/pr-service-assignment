package web

import (
	"net/http"

	"github.com/guarref/pr-service-assignment/internal/resperrors"
	"github.com/guarref/pr-service-assignment/internal/service"
	"github.com/guarref/pr-service-assignment/internal/web/odomains"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{service: s}
}

// POST /users/setIsActive
func (h *UserHandler) PostUsersSetIsActive(ctx echo.Context) error {
	var body odomains.PostUsersSetIsActiveJSONRequestBody

	if err := ctx.Bind(&body); err != nil {
		resp := NewErrorResponse(odomains.NOTFOUND, "invalid request body")
		return ctx.JSON(http.StatusBadRequest, resp)
	}

	if body.UserId == "" {
		return writeDomainError(ctx, resperrors.ErrBadRequest)
	}

	user, err := h.service.SetFlagIsActive(ctx.Request().Context(), body.UserId, body.IsActive)
	if err != nil {
		return writeDomainError(ctx, err)
	}

	respUser := toOAPIUser(user)

	return ctx.JSON(http.StatusOK, struct {
		User odomains.User `json:"user"`
	}{
		User: respUser,
	})
}