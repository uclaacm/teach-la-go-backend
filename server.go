package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/uclaacm/teach-la-go-backend/db"

	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

// DEFAULTPORT to serve on.
const DEFAULTPORT = "8081"

func serve(c *cli.Context) error {
	e := echo.New()
	e.HideBanner = true

	if c.Bool("verbose") {
		e.Logger.SetLevel(log.DEBUG)
	}

	// middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderContentType},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}))

	// try to set up firestore connection.
	// check the following resources in order:
	// * $TLACFG environment variable
	// * creds.json file (or other)
	envCfg, jsonPath := c.String("env-var"), c.String("config-path")
	var (
		d   *db.DB
		err error
	)
	switch {
	case envCfg != "":
		d, err = db.Open(context.Background(), os.Getenv(envCfg))
	case jsonPath != "":
		d, err = db.OpenFromJSON(context.Background(), jsonPath)
	default:
		return fmt.Errorf("failed to locate credentials")
	}
	if err != nil {
		e.Logger.Fatal(errors.Wrap(err, "failed to open connection to firestore"))
	}
	defer d.Close()

	// user management
	e.GET("/user/get", d.GetUser)
	e.PUT("/user/update", d.UpdateUser)
	e.POST("/user/create", d.CreateUser)
	e.POST("/user/classes", d.GetClasses)

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

	return nil
}

func main() {
	app := &cli.App{
		Name:                 "Teach LA Go Backend",
		Description:          "Binary application for Teach LA's editor backend!",
		Version:              "1.0.0",
		HideHelpCommand:      true,
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "verbose",
				Value: false,
				Usage: "Enable verbosity",
			},
			&cli.IntFlag{
				Name:     "port",
				Aliases:  []string{"p"},
				Required: false,
				Value:    8081,
				Usage:    "Port to serve the backend on",
			},
			&cli.StringFlag{
				Name:     "config-path",
				Aliases:  []string{"c"},
				Required: false,
				Value:    "creds.json",
				Usage:    "Specify a path to JSON Firebase credentials",
			},
			&cli.StringFlag{
				Name:     "env-var",
				Aliases:  []string{"ev"},
				Required: false,
				Value:    db.DefaultEnvVar,
				Usage:    "Specify an alternative environment variable to scan for credentials",
			},
		},
		Action: serve,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("Failed to start! %v", err)
	}
}
