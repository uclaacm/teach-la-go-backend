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
	Target string `json:"target"`
	Body   string `json:"body"`
}
type StringSet map[string]bool
type Session struct {
	// Map UIDs to their websocket.Conn
	Conns   map[string]*websocket.Conn
	Teacher string
	// Map UIDs to the other UIDs they send messages to
	SendTo map[string]StringSet
	Lock   sync.Mutex
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
	s.SendTo[uid] = make(StringSet)
	return nil
}

func (s *Session) RemoveConn(uid string) error {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	if s.Conns[uid] != nil {
		return errors.New("Could not remove unconnected user")
	}
	delete(s.Conns, uid)
	delete(s.SendTo, uid)
	// Stop other connections from sending to removed connection
	for _, v := range s.SendTo {
		delete(v, uid)
	}
	return nil
}

func (s *Session) BroadcastAll(msg Message) error {
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

func (s *Session) BroadcastError(uid string, err string) error {
	errorMsg := Message{
		Author: uid,
		Type:   "ERROR",
		Body:   err,
	}
	if broadcastErr := s.BroadcastTo(errorMsg, uid); broadcastErr != nil {
		return broadcastErr
	}
	return nil
}

func (s *Session) BroadcastTo(msg Message, uids ...string) error {
	var mostRecentError error
	s.Lock.Lock()
	defer s.Lock.Unlock()
	for _, uid := range uids {
		if err := websocket.JSON.Send(s.Conns[uid], msg); err != nil {
			mostRecentError = err
		}
	}
	return mostRecentError
}

func (s *Session) BroadcastToSet(msg Message, uids StringSet) error {
	var mostRecentError error
	s.Lock.Lock()
	defer s.Lock.Unlock()
	for uid := range uids {
		if err := websocket.JSON.Send(s.Conns[uid], msg); err != nil {
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
		Conns:  make(map[string]*websocket.Conn),
		SendTo: make(map[string]StringSet),
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
					for k := range session.Conns {
						session.Teacher = k
					}
				}
			}()
		}

		for {
			var msg Message

			if err := websocket.JSON.Receive(ws, &msg); err != nil {
				if err2 := session.BroadcastError(uid, err.Error()); err2 != nil {
					break
				}

				if removeErr := session.RemoveConn(uid); removeErr != nil {
					break
				}

				break
			}

			switch msg.Type {
			case "READ":
				if msg.Author == session.Teacher {
					if studentSendTo, ok := session.SendTo[msg.Target]; ok {
						studentSendTo[uid] = true
					} else {
						if err := session.BroadcastError(uid, "Student does not exist"); err != nil {
							break
						}
					}
				} else {
					if err := session.BroadcastError(uid, "Teacher permission required to request access to student"); err != nil {
						break
					}
				}
			default:
				if err := session.BroadcastToSet(msg, session.SendTo[uid]); err != nil {
					break
				}
			}
		}
	}).ServeHTTP(c.Response(), c.Request())

	return nil
}
