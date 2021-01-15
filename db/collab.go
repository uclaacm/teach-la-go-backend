package db

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uclaacm/teach-la-go-backend/httpext"
	"golang.org/x/net/websocket"
)

// Session describes a collaborative coding environment.
type Session struct {
	// Map UIDs to their websocket.Conn
	Conns map[string]*websocket.Conn
}

// Maps session IDs to Session object
var sessions map[string]Session
var sessionsLock sync.Mutex

func init() {
	sessions = make(map[string]Session)
}

func (d *DB) CreateCollab(c echo.Context) error {
	var body struct {
		UID string `json:"uid"`
	}
	if err := httpext.RequestBodyTo(c.Request(), &body); err != nil {
		return c.String(http.StatusInternalServerError, "failed to read request body")
	}

	sessionId := uuid.New().String()
	sessions[sessionId] = Session{
		Conns: make(map[string]*websocket.Conn),
	}

	// Kill session if no connections every minute
	go func() {
		ticker := time.NewTicker(time.Minute)
		for range ticker.C {
			fmt.Printf("%+v\n", sessions)
			sessionsLock.Lock()
			if session, ok := sessions[sessionId]; ok && len(session.Conns) == 0 {
				fmt.Printf("Deleting session")
				delete(sessions, sessionId)
				sessionsLock.Unlock()
				ticker.Stop()
				return
			}
			sessionsLock.Unlock()
		}
	}()

	return c.String(http.StatusCreated, sessionId)
}
