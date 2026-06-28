package middlewares

import (
	"net/http"

	"base-api/opt/db"
	"base-api/pkg/ctxdb"

	"github.com/labstack/echo/v4"
)

// TenantDBContextKey is the key used to store the tenant DB in echo.Context
const TenantDBContextKey = "tenant_db"

// TenantDBResolver creates a middleware that resolves tenant database connections
// based on the X-Tenant-Slug header and injects them into both echo.Context
// and the request's context.Context.
func TenantDBResolver(poolMgr *database.TenantPoolManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			slug := c.Request().Header.Get("X-Tenant-Slug")
			if slug == "" {
				return echo.NewHTTPError(http.StatusBadRequest, "tenant slug required")
			}

			db, err := poolMgr.GetOrCreate(c.Request().Context(), slug)
			if err != nil {
				return echo.NewHTTPError(http.StatusNotFound, "tenant database not found")
			}

			// Store in echo.Context for direct handler access
			c.Set(TenantDBContextKey, db)

			// Store in request's context.Context for use cases and repositories
			ctx := ctxdb.WithDB(c.Request().Context(), db)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
