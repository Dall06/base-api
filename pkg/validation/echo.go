package validation

import (
	"base-api/pkg/errs"

	"github.com/labstack/echo/v4"
)

// BindAndValidate binds the request body to the struct and validates it.
// Returns a ValueError with user-friendly messages if validation fails.
func BindAndValidate(ctx echo.Context, req interface{}) error {
	if err := ctx.Bind(req); err != nil {
		return errs.ValueError("bad request")
	}
	if err := Struct(req); err != nil {
		return errs.ValueError("%s", err.Error())
	}
	return nil
}
