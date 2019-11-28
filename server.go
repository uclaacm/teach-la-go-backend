package main

import (
	"log"
	"net/http"
	"time"

	"./lib"
	"./dbTools"
	m "./middleware"
)

// PORT defines where we serve the backend.
const PORT = ":8081"

func main() {
	// acquire firestore client.
	// fails early if we cannot acquire one.
	client := dbTools.GetDB()
	defer client.Close()

	// establish handlers.
	userMgr := lib.HandleUsers{Client: client}
	progMgr := lib.HandlePrograms{Client: client}

	log.Printf("successfully initialized firestore client and route handlers")

	// set up basic server
	router := http.NewServeMux()
	log.Printf("server initialized.")

	// user management
	router.Handle("/userData/", m.WithCORS(userMgr))

	// program management
	router.Handle("/programs/", m.WithCORS(progMgr))

	// fallback route
	router.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusNotFound)
	})

	log.Printf("endpoints initialized, serving.")

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
