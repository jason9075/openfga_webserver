package handler

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

// PublicHandler: 公開頁面
func PublicHandler(c echo.Context) error {
	return c.String(http.StatusOK, "這是公開頁面，任何人都可以存取。")
}

// PageHandler: 頁面
func PageHandler(c echo.Context) error {
	userUri := c.Param("page-uri")
	role := c.QueryParam("access")
	return c.String(http.StatusOK, fmt.Sprintf("Hello, user %v! This is %v's page.", role, userUri))
}
