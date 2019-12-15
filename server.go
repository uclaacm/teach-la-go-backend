package main

import (
	"log"
	"net/http"
	"time"

	"./lib"
	m "./middleware"
	"./tools"
)

// PORT defines where we serve the backend.
const PORT = ":8081"

func main() {
	// acquire firestore client.
	// fails early if we cannot acquire one.
	client := tools.GetDB()
	defer client.Close()

	// establish handlers.
	userMgr := lib.HandleUsers{Client: client}
	progMgr := lib.HandlePrograms{Client: client}

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

	log.Printf("serving on %s", PORT)

	// finally, serve the backend
	log.Fatal(s.ListenAndServe())
}
