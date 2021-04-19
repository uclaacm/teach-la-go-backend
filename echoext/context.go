package echoext

import (
	"github.com/uclaacm/teach-la-go-backend/db"

	"github.com/labstack/echo/v4"
)

type DBContext struct {
	echo.Context
	db.TLADB	
}
