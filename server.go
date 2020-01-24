package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"./lib"
	m "./middleware"
)

// PORT defines where we serve the backend.
const PORT = ":8081"

func main() {
	// set up context for main routine.
	mainContext := context.Background()

	// acquire DB client.
	// fails early if we cannot acquire one.
	var (
		d *lib.DB
		err error
	)
	if d, err = lib.OpenFromEnv(context.Background()); err != nil {
		log.Fatalf("failed to open DB client. %s", err)
	}
	defer d.Close()

	// establish handlers.
	userMgr := lib.HandleUsers{Client: d.Client}
	progMgr := lib.HandlePrograms{Client: d.Client}

	log.Printf("initialized firestore client and route handlers.")

	// set up multiplexer.
	router := http.NewServeMux()
	log.Printf("multiplexer initialized.")

	// user management
	userCORS := m.CORSConfig{
		AllowedHeaders: []string{"Content-Type"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut},
	}
	router.Handle("/userData/", m.WithCORSConfig(userMgr, userCORS))

	// program management
	progCORS := m.CORSConfig{
		AllowedHeaders: []string{"Content-Type"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}
	router.Handle("/programs/", m.WithCORSConfig(progMgr, progCORS))

	// fallback route
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusNotFound)
	})

	log.Printf("endpoints initialized.")

	// server configuration
	s := &http.Server{
		Addr:           PORT,
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
		log.Printf("serving on %s", PORT)
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
