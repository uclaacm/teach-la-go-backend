package lib

import (
	"net/http"

	"github.com/labstack/echo"
)

func (h *Handler) CreateProgram(c echo.Context) error {
	return c.String(http.StatusNotImplemented, "")
}

func (h *Handler) UpdatePrograms(c echo.Context) error {
	return c.String(http.StatusNotImplemented, "")
}

func (h *Handler) DeletePrograms(c echo.Context) error {
	return c.String(http.StatusNotImplemented, "")
}

func (h *Handler) GetProgram(c echo.Context) error {
	return c.String(http.StatusNotImplemented, "")
}
