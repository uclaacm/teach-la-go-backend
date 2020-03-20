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

func main() {
	// set up context for main routine.
	mainContext := context.Background()

	// check for PORT variable.
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalf("no $PORT environment variable provided.")
	}

	// acquire DB client.
	// fails early if we cannot acquire one.
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
	router.HandleFunc("/user/get", m.WithCORSConfig(d.HandleGetUser, m.CORSConfig{
		AllowedOrigins: []string{"http://localhost:8080", "editor.uclaacm.com"},
		AllowedMethods: []string{http.MethodGet},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         3200,
	}))
	router.HandleFunc("/user/update", m.WithCORSConfig(d.HandleUpdateUser, m.CORSConfig{
		AllowedOrigins: []string{"http://localhost:8080", "editor.uclaacm.com"},
		AllowedMethods: []string{http.MethodPut},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         3200,
	}))
	router.HandleFunc("/user/create", m.WithCORSConfig(d.HandleInitializeUser, m.CORSConfig{
		AllowedOrigins: []string{"http://localhost:8080", "editor.uclaacm.com"},
		AllowedMethods: []string{http.MethodPost},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         3200,
	}))

	// program management
	router.HandleFunc("/program/get", m.WithCORSConfig(d.HandleGetProgram, m.CORSConfig{
		AllowedOrigins: []string{"http://localhost:8080", "editor.uclaacm.com"},
		AllowedMethods: []string{http.MethodGet},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         3200,
	}))
	router.HandleFunc("/program/update", m.WithCORSConfig(d.HandleUpdateProgram, m.CORSConfig{
		AllowedOrigins: []string{"http://localhost:8080", "editor.uclaacm.com"},
		AllowedMethods: []string{http.MethodPut},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         3200,
	}))
	router.HandleFunc("/program/create", m.WithCORSConfig(d.HandleInitializeProgram, m.CORSConfig{
		AllowedOrigins: []string{"http://localhost:8080", "editor.uclaacm.com"},
		AllowedMethods: []string{http.MethodPost},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         3200,
	}))
	router.HandleFunc("/program/delete", m.WithCORSConfig(d.HandleDeleteProgram, m.CORSConfig{
		AllowedOrigins: []string{"http://localhost:8080", "editor.uclaacm.com"},
		AllowedMethods: []string{http.MethodDelete},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         3200,
	}))

	// fallback route
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusNotFound)
	})

	log.Printf("endpoints initialized.")

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
	timeout, terminate := context.WithDeadline(mainContext, time.Now().Add(10*time.Second))
	defer terminate()
	s.Shutdown(timeout)
}
