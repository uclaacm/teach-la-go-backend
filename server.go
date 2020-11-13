package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
	"strconv"
	"io/ioutil"

	"github.com/uclaacm/teach-la-go-backend/db"

	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"github.com/joho/godotenv"
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

	CfgPath := c.String("cred")
	cfg := ""

	// if env-cred is NOT set, and json- cred is set, open json credentials
	// otherwise, open env credentials
	if !c.Bool("env-cred") && c.Bool("json-cred") {
		bytes, err := ioutil.ReadFile(CfgPath + "cred.json")
		if err != nil {
			e.Logger.Fatal(errors.Wrap(err, "failed to open .json file"))
		}
		cfg = string(bytes)
		if cfg == "" {
			e.Logger.Fatalf("no config provided", db.DefaultEnvVar)
		}
	} else {
		if err := godotenv.Load(CfgPath + ".env"); err != nil {
			e.Logger.Fatal(errors.Wrap(err, "failed to open .env file"))
		}
		cfg = os.Getenv(db.DefaultEnvVar)
		if cfg == "" {
			e.Logger.Fatalf("no $%s environment variable provided", db.DefaultEnvVar)
		}
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
	port := c.Int("port")

	// server configuration
	s := &http.Server{
		Addr:           ":" + strconv.Itoa(port),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	e.Logger.Fatal(e.StartServer(s))

	return nil
}

func main() {
	app := &cli.App{
		Name: "Teach LA Go Backend",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "creds",
				Aliases: []string{"c"},
				Value:   "./",
				Usage:   "Provide a path to the database credentials being used (defaults to `.env` and `*.json`, in that order).",
			},
			&cli.BoolFlag{
				Name:    "env-cred",
				Aliases: []string{"e"},
				Value:   false,
				Usage:   "use an env file to specify credentias",
			},
			&cli.BoolFlag{
				Name:    "json-cred",
				Aliases: []string{"j"},
				Value:   false,
				Usage:   "use a JSON file to specify credentials",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Value:   false,
				Usage:   "Change the log level used by echo's logger middleware.",
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   8081,
				Usage:   "Change the port number",
			},
			
		},
		Action: serve,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("Failed to start! %v", err)
	}
}
