package middlewares

import (
	"base-api/opt/guardian"
	"base-api/pkg/auth"

	"github.com/labstack/echo/v4"
)

// NewMemberAuth creates a JWT authentication middleware for the member portal.
// Reads token from the bro_member_token HttpOnly cookie (separate from staff cookie).
// Validates JWT signature + session store, and enforces role == "member".
func NewMemberAuth(g *guardian.Guardian) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			tokenString := auth.GetMemberTokenFromCookie(ctx)
			if tokenString == "" {
				return echo.NewHTTPError(401, "unauthorized")
			}

			claims, err := g.ValidateAccessToken(ctx.Request().Context(), tokenString)
			if err != nil {
				return echo.NewHTTPError(401, "unauthorized")
			}

			if claims.Role != guardian.RoleMember {
				return echo.NewHTTPError(403, "forbidden: member access only")
			}

			ctx.Set("member_id", claims.MemberID)
			ctx.Set("slug", claims.Slug)
			ctx.Set("role", claims.Role)
			ctx.Set("jti", claims.ID)

			return next(ctx)
		}
	}
}
