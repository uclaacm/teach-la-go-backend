package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/uclaacm/teach-la-go-backend/db"

	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
)

// DEFAULTPORT to serve on.
const DEFAULTPORT = "8081"

func main() {
	e := echo.New()
	e.HideBanner = true

	// middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://editor.uclaacm.com", "http://localhost:8080"},
		AllowHeaders: []string{echo.HeaderContentType},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}))

	// try to set up firestore connection through env vars
	cfg := os.Getenv(db.DefaultEnvVar)
	if cfg == "" {
		e.Logger.Fatalf("no $%s environment variable provided", db.DefaultEnvVar)
	}
	d, err := db.Open(context.Background(), cfg)
	if err != nil {
		e.Logger.Fatal(errors.Wrap(err, "failed to open connection to firestore"))
	}
	defer d.Close()

	// user management
	e.GET("/user/get", d.GetUser)
	e.PUT("/user/update", d.UpdateUser)
	e.POST("/user/create", d.CreateUser)

	// program management
	e.GET("/program/get", d.GetProgram)
	e.PUT("/program/update", d.UpdateProgram)
	e.POST("/program/create", d.CreateProgram)
	e.DELETE("/program/delete", d.DeleteProgram)

	//class management
	e.GET("/class/get", d.GetClass)
	e.POST("/class/create", d.CreateClass)
	e.PUT("/class/join", d.JoinClass)
	e.PUT("/class/leave", d.LeaveClass)

	// check for PORT variable.
	port := os.Getenv("PORT")
	if port == "" {
		e.Logger.Debugf("no $PORT environment variable provided, defaulting to '%s'", DEFAULTPORT)
		port = DEFAULTPORT
	}

	// server configuration
	s := &http.Server{
		Addr:           ":" + port,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	e.Logger.Fatal(e.StartServer(s))
}
