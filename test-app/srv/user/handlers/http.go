package handlers

import (
	"net/http"
	"strings"

	"github.com/diegoaleon/test-app/pkg/errs"
	"github.com/diegoaleon/test-app/pkg/jwt"
	"github.com/diegoaleon/test-app/srv/user/ports"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	usecase      ports.UserUsecase
	jwtGenerator *jwt.Generator
}

func NewUserHandler(usecase ports.UserUsecase, jwtGenerator *jwt.Generator) ports.UserHandler {
	return &UserHandler{
		usecase:      usecase,
		jwtGenerator: jwtGenerator,
	}
}

func (h *UserHandler) RegisterRoutes(e *echo.Echo) {
	api := e.Group("/api/v1")

	api.POST("/auth/signup", h.Signup)
	api.POST("/auth/login", h.Login)
	api.GET("/users/me", h.GetMe, h.AuthMiddleware)
}

func (h *UserHandler) Signup(c echo.Context) error {
	var input ports.SignupInput
	if err := c.Bind(&input); err != nil {
		return errs.Handle(c, errs.ValueError("cuerpo de petición inválido"))
	}

	res, err := h.usecase.Signup(c.Request().Context(), input)
	if err != nil {
		return errs.Handle(c, err)
	}

	return c.JSON(http.StatusCreated, res)
}

func (h *UserHandler) Login(c echo.Context) error {
	var input ports.LoginInput
	if err := c.Bind(&input); err != nil {
		return errs.Handle(c, errs.ValueError("cuerpo de petición inválido"))
	}

	res, err := h.usecase.Login(c.Request().Context(), input)
	if err != nil {
		return errs.Handle(c, err)
	}

	return c.JSON(http.StatusOK, res)
}

func (h *UserHandler) GetMe(c echo.Context) error {
	userID := c.Get("user_id").(string)
	user, err := h.usecase.GetByID(c.Request().Context(), userID)
	if err != nil {
		return errs.Handle(c, err)
	}

	return c.JSON(http.StatusOK, user)
}

func (h *UserHandler) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return errs.Handle(c, errs.UnauthorizedError("token faltante"))
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return errs.Handle(c, errs.UnauthorizedError("formato de token inválido"))
		}

		claims, err := h.jwtGenerator.Validate(parts[1])
		if err != nil {
			return errs.Handle(c, errs.UnauthorizedError("token inválido o expirado"))
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		return next(c)
	}
}
