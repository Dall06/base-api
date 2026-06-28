package middlewares

import (
	"net/http"

	"github.com/diegoaleon/test-app/pkg/errs"

	"github.com/labstack/echo/v4"
)

// GatewayTenantResolver extracts and validates the tenant slug from the URL.
// Used by the gateway to inject X-Tenant-Slug header for downstream services.
func GatewayTenantResolver() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			slug := ctx.Param("slug")
			if slug == "" {
				return ctx.JSON(http.StatusBadRequest, errs.Response{Error: "tenant slug is required"})
			}

			// Validate slug format (basic validation)
			if len(slug) < 3 || len(slug) > 50 {
				return ctx.JSON(http.StatusBadRequest, errs.Response{Error: "invalid tenant slug"})
			}

			// Store slug in context for downstream handlers
			ctx.Set("tenant_slug", slug)

			// Add slug to request headers for downstream services
			ctx.Request().Header.Set("X-Tenant-Slug", slug)

			return next(ctx)
		}
	}
}
