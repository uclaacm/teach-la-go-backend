package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"./db"
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
