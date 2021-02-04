package db

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/uclaacm/teach-la-go-backend/httpext"
	"golang.org/x/net/websocket"
)

// Session describes a collaborative coding environment.
type Message struct {
	Author string
	Type   string
	Body   string
}
type Session struct {
	// Map UIDs to their websocket.Conn
	Conns map[string]*websocket.Conn
	Lock  sync.Mutex
}

// Maps session IDs to Session object
var (
	sessions     map[string]Session
	sessionsLock sync.Mutex
)

func init() {
	sessions = make(map[string]Session)
}

func (s *Session) AddConn(uid string, conn *websocket.Conn) error {
	s.Lock.Lock()
	if s.Conns[uid] != nil {
		s.Lock.Unlock()
		return errors.New("User is already connected")
	}
	s.Conns[uid] = conn
	s.Lock.Unlock()
	return nil
}

func (s *Session) RemoveConn(uid string) error {
	s.Lock.Lock()
	if s.Conns[uid] != nil {
		s.Lock.Unlock()
		return errors.New("Could not remove unconnected user")
	}
	delete(s.Conns, uid)
	s.Lock.Unlock()
	return nil
}

func (s *Session) Broadcast(msg Message) error {
	var mostRecentError error = nil
	s.Lock.Lock()
	for _, conn := range s.Conns {
		if err := websocket.JSON.Send(conn, msg); err != nil {
			mostRecentError = err
		}
	}
	s.Lock.Unlock()
	return mostRecentError
}

func (d *DB) CreateCollab(c echo.Context) error {
	var body struct {
		UID string `json:"uid"`
	}
	if err := httpext.RequestBodyTo(c.Request(), &body); err != nil {
		return c.String(http.StatusInternalServerError, "failed to read request body")
	}

	sessionId := uuid.New().String()
	sessionsLock.Lock()
	sessions[sessionId] = Session{
		Conns: make(map[string]*websocket.Conn),
	}
	sessionsLock.Unlock()

	// Kill session if no connections every minute
	go func() {
		ticker := time.NewTicker(time.Minute)
		for range ticker.C {
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

func (d *DB) JoinCollab(c echo.Context) error {
	sessionId := c.Param("id")
	uid := uuid.New().String() // What will we be using as identifiers?
	session, ok := sessions[sessionId]

	if !ok {
		return c.String(http.StatusNotFound, "Session does not exist.")
	}

	websocket.Handler(func(ws *websocket.Conn) {
		if err := session.AddConn(uid, ws); err != nil {
			return
		}

		for {
			var msg Message

			if err := websocket.JSON.Receive(ws, &msg); err != nil {
				errorMsg := Message{
					Author: uid,
					Type:   "ERROR",
					Body:   err.Error(),
				}
				if broadcastErr := session.Broadcast(errorMsg); broadcastErr != nil {
					break
				}

				if removeErr := session.RemoveConn(uid); removeErr != nil {
					break
				}

				break
			}

			if broadcastErr := session.Broadcast(msg); broadcastErr != nil {
				break
			}
		}
	}).ServeHTTP(c.Response(), c.Request())

	return nil
}
