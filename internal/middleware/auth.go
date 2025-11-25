package middleware

import (
	"github.com/naozine/project_crud_with_auth_tmpl/internal/appcontext"

	"github.com/labstack/echo/v4"
	"github.com/naozine/nz-magic-link/magiclink"
)

func UserContextMiddleware(ml *magiclink.MagicLink) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userEmail, isLoggedIn := ml.GetUserID(c)

			var hasPasskey bool
			if isLoggedIn {
				creds, err := ml.DB.GetPasskeyCredentialsByUserID(userEmail)
				if err == nil && len(creds) > 0 {
					hasPasskey = true
				}
			}

			// Set user info to request context
			ctx := c.Request().Context()
			ctx = appcontext.WithUser(ctx, userEmail, isLoggedIn, hasPasskey)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
