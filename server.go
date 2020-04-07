package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/uclaacm/teach-la-go-backend/db"
	m "github.com/uclaacm/teach-la-go-backend/middleware"
)

// DEFAULTPORT to serve on.
const DEFAULTPORT = "8081"

func main() {
	var (
		d   *db.DB
		err error
	)

	if d, err = db.OpenFromEnv(context.Background()); err != nil {
		log.Fatalf("failed to open DB client. %s", err)
	}
	defer d.Close()
	log.Printf("initialized database client")

	// set up multiplexer.
	router := http.NewServeMux()

	// user management
	router.HandleFunc("/user/get", d.HandleGetUser)
	router.HandleFunc("/user/update", d.HandleUpdateUser)
	router.HandleFunc("/user/create", d.HandleInitializeUser)

	// program management
	router.HandleFunc("/program/get", d.HandleGetProgram)
	router.HandleFunc("/program/update", d.HandleUpdateProgram)
	router.HandleFunc("/program/create", d.HandleInitializeProgram)
	router.HandleFunc("/program/delete", d.HandleDeleteProgram)

	//class management
	router.HandleFunc("/class/create", d.HandleCreateClass)
	router.HandleFunc("/class/get", d.HandleGetClass)
	router.HandleFunc("/class/join", d.HandleJoinClass)
	router.HandleFunc("/class/leave", d.HandleLeaveClass)

	// fallback route
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusNotFound)
	})

	log.Printf("endpoints initialized.")

	// check for PORT variable.
	port := os.Getenv("PORT")
	if port == "" {
		log.Printf("no $PORT environment variable provided, defaulting to '%s'", DEFAULTPORT)
		port = "8081"
	}

	// server configuration
	s := &http.Server{
		Addr:           ":" + port,
		Handler:        m.LogRequest(router),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// serve backend via anonymous goroutine, cancelling on
	// system interrupt.
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt)
	go func() {
		log.Printf("serving on :%s", port)
		log.Fatal(s.ListenAndServe())
	}()

	// wait for system interrupt to call shutdown on the server.
	<-kill
	log.Printf("received kill signal, attempting to gracefully shut down.")

	// server has 10 seconds from interrupt to gracefully shutdown.
	timeout, terminate := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer terminate()
	s.Shutdown(timeout)
}
