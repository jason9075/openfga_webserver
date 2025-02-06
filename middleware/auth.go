package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/openfga/go-sdk/client"
)

type OpenFGAConfig struct {
	Client *client.OpenFgaClient
}

func Authorization(config OpenFGAConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 假設我們使用 FakeAccessMiddleware() 放入 "userRole"
			role := c.QueryParam("access")
			if role == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing role in context")
			}

			userUri := c.Param("page-uri")

			// Create the request body
			body := client.ClientCheckRequest{
				User:     fmt.Sprintf("user:%s", role),
				Relation: "can_view",
				Object:   fmt.Sprintf("page:%s", userUri),
			}
			fmt.Println("body: ", body)

			// Check if the user is allowed to access the page
			resp, err := config.Client.Check(context.Background()).
				Body(body).
				Execute()
			if err != nil {
				fmt.Println("Error: ", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}

			if !resp.GetAllowed() {
				return echo.NewHTTPError(http.StatusForbidden, "Access denied")
			}

			return next(c)
		}
	}
}
