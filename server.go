package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	"github.com/uclaacm/teach-la-go-backend/db"
	"github.com/uclaacm/teach-la-go-backend/handler"
	"github.com/urfave/cli/v2"
)

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

	// Check for working credentials in the following partial order:
	// - JSON
	// - .env
	// - TLACFG
	jsonPath, dotenvPath := c.String("json"), c.String("dotenv")
	var (
		d   *db.DB
		err error
	)
	switch {
	case jsonPath != "":
		d, err = db.OpenFromJSON(context.Background(), jsonPath)
	case dotenvPath != "":
		if err := godotenv.Load(dotenvPath); err != nil {
			e.Logger.Error(errors.Wrap(err, "failed to open .env file"))
		}
		d, err = db.Open(context.Background(), os.Getenv(db.DefaultEnvVar))
	default:
		d, err = db.Open(context.Background(), os.Getenv(db.DefaultEnvVar))
	}
	if err != nil {
		e.Logger.Fatal(errors.Wrap(err, "failed to open connection to firestore"))
		return err
	}
	defer d.Close()

	// Register our database handler to every Echo context.
	e.Use(func(nxt echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return nxt(&db.DBContext{
				Context: c,
				TLADB:   d,
			})
		}
	})

	// user management
	e.GET("/user/get", handler.GetUser)
	e.PUT("/user/update", d.UpdateUser)
	e.POST("/user/create", d.CreateUser)

	// program management
	e.GET("/program/get", d.GetProgram)
	e.PUT("/program/update", d.UpdateProgram)
	e.POST("/program/create", d.CreateProgram)
	e.DELETE("/program/delete", d.DeleteProgram)

	// class management
	e.POST("/class/get", d.GetClass)
	e.POST("/class/create", d.CreateClass)
	e.PUT("/class/join", d.JoinClass)
	e.PUT("/class/leave", d.LeaveClass)
	e.POST("/class/members", d.GetClassMembers)

	// collaborative coding management
	e.POST("/collab/create", d.CreateCollab)
	e.GET("/collab/join/:id", d.JoinCollab)

	// check for PORT variable.
	var port string
	if osPort := os.Getenv("PORT"); osPort != "" {
		port = osPort
	} else {
		port = c.String("port")
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
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "Print the version and exit",
	}
	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h"},
		Usage:   "Show help",
	}

	app := &cli.App{
		Name:                 "Teach LA Go Backend",
		Usage:                "tlabe [options]",
		Description:          "Teach LA's editor backend.",
		Version:              "1.0.0",
		HideHelpCommand:      true,
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "dotenv",
				Aliases:  []string{"e"},
				Required: false,
				Usage:    "Specify a path to a dotenv file to specify credentials",
			},
			&cli.StringFlag{
				Name:     "json",
				Aliases:  []string{"j"},
				Required: false,
				Usage:    "Specify a path to a JSON file to specify credentials",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Value:   false,
				Usage:   "Change the log level used by echo's logger middleware",
			},
			&cli.StringFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   "8081",
				Usage:   "Change the port number",
			},
		},
		Action: serve,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("Failed to start! %v", err)
	}
}
