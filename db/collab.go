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
	Author string `json:"author"`
	Type   string `json:"type"`
	Body   string `json:"body"`
}
type Session struct {
	// Map UIDs to their websocket.Conn
	Conns   map[string]*websocket.Conn
	Teacher string
	Lock    sync.Mutex
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
	defer s.Lock.Unlock()
	if s.Conns[uid] != nil {
		return errors.New("User is already connected")
	}
	s.Conns[uid] = conn
	return nil
}

func (s *Session) RemoveConn(uid string) error {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	if s.Conns[uid] != nil {
		return errors.New("Could not remove unconnected user")
	}
	delete(s.Conns, uid)
	return nil
}

func (s *Session) Broadcast(msg Message) error {
	var mostRecentError error
	s.Lock.Lock()
	defer s.Lock.Unlock()
	for _, conn := range s.Conns {
		if err := websocket.JSON.Send(conn, msg); err != nil {
			mostRecentError = err
		}
	}
	return mostRecentError
}

// CreateCollab creates a collaborative session, setting up the session's websocket.
// Request Body:
// {
//    uid: UID for the user the program belongs to
//    name: optional name identifier for the session, defaults to random UUID.
// }
//
// Returns status 201 created on success.
func (d *DB) CreateCollab(c echo.Context) error {
	var body struct {
		Name string `json:"name"`
		UID  string `json:"uid"`
	}
	if err := httpext.RequestBodyTo(c.Request(), &body); err != nil {
		return c.String(http.StatusInternalServerError, "failed to read request body")
	}

	sessionId := uuid.New().String()
	if body.Name != "" {
		sessionId = body.Name
	}

	if _, ok := sessions[sessionId]; ok {
		return c.String(http.StatusBadRequest, "session with same name already exists")
	}

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
		if len(session.Conns) == 1 {
			session.Teacher = uid
			defer func() {
				if len(session.Conns) >= 1 {
					// If teacher leaves, choose arbitrary member to be teacher
					for k, _ := range session.Conns {
						session.Teacher = k
					}
				}
			}()
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
