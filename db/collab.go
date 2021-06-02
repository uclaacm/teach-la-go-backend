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

type Message struct {
	Author string `json:"author"`
	Type   string `json:"type"`
	Target string `json:"target"`
	Body   string `json:"body"`
}
type stringSet map[string]bool

// Session describes a collaborative coding environment.
type Session struct {
	*sync.Mutex

	// Map UIDs to their websocket.Conn
	Conns   map[string]*Connection
	Teacher string
}
type Connection struct {
	*websocket.Conn
	UID string
	// Set of UIDs which are notified when this connection has changes
	Subscriptions stringSet
}

// Maps session IDs to Session object
var (
	sessions sync.Map
)

func (s *Session) AddConn(uid string, conn *websocket.Conn) error {
	s.Lock()
	defer s.Unlock()
	if s.Conns[uid] != nil {
		return errors.New("User is already connected")
	}
	connection := &Connection{Conn: conn, UID: uid, Subscriptions: make(stringSet)}
	s.Conns[uid] = connection
	return nil
}

func (s *Session) RemoveConn(uid string) error {
	s.Lock()
	defer s.Unlock()
	if s.Conns[uid] != nil {
		return errors.New("Could not remove unconnected user")
	}
	delete(s.Conns, uid)
	// Stop other connections from sending to removed connection
	for _, conn := range s.Conns {
		delete(conn.Subscriptions, uid)
	}
	return nil
}

// BroadcastAll sends a message to all active connections
func (s *Session) BroadcastAll(msg Message) (lastErr error) {
	s.Lock()
	defer s.Unlock()
	for _, conn := range s.Conns {
		if err := websocket.JSON.Send(conn.Conn, msg); err != nil {
			lastErr = err
		}
	}
	return
}

// BroadcastError creates and sends an Error message given a string err
// to a given uid
func (s *Session) BroadcastError(uid string, err string) error {
	errorMsg := Message{
		Author: uid,
		Type:   msgTypeError,
		Body:   err,
	}
	if broadcastErr := s.BroadcastTo(errorMsg, uid); broadcastErr != nil {
		return broadcastErr
	}
	return nil
}

// BroadcastTo sends a Message msg to the connections associated with the
// provided uids
func (s *Session) BroadcastTo(msg Message, uids ...string) (lastErr error) {
	s.Lock()
	defer s.Unlock()
	for _, uid := range uids {
		if err := websocket.JSON.Send(s.Conns[uid].Conn, msg); err != nil {
			lastErr = err
		}
	}
	return
}

// BroadcastToSet sends a Message msg to the connections associated with the
// provided uids in a stringSet
func (s *Session) BroadcastToSet(msg Message, uids stringSet) (lastErr error) {
	s.Lock()
	defer s.Unlock()
	for uid := range uids {
		if err := websocket.JSON.Send(s.Conns[uid].Conn, msg); err != nil {
			lastErr = err
		}
	}
	return
}

func (s *Session) RequestAccess(uid string, msg Message) error {
	s.Lock()
	defer s.Unlock()
	if msg.Author == s.Teacher {
		if conn, ok := s.Conns[msg.Target]; ok {
			conn.Subscriptions[uid] = true
		} else {
			return s.BroadcastError(uid, "Student does not exist")
		}
	} else {
		return s.BroadcastError(uid, "Teacher permission required to request access to student")
	}
	return nil
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

	sessionID := uuid.New().String()
	if body.Name != "" {
		sessionID = body.Name
	}

	if _, ok := sessions.Load(sessionID); ok {
		return c.String(http.StatusBadRequest, "session with same name already exists")
	}

	sessions.Store(sessionID, Session{
		Conns: make(map[string]*Connection),
	})

	// Kill session if no connections every minute
	go func() {
		ticker := time.NewTicker(time.Minute)
		for range ticker.C {
			if sessionIFace, ok := sessions.Load(sessionID); ok && len(sessionIFace.(Session).Conns) == 0 {
				fmt.Printf("Deleting session")
				sessions.Delete(sessionID)
				ticker.Stop()
				return
			}
		}
	}()

	return c.String(http.StatusCreated, sessionID)
}

func (d *DB) JoinCollab(c echo.Context) error {
	sessionID := c.Param("id")
	uid := uuid.New().String() // What will we be using as identifiers?
	sessionIFace, ok := sessions.Load(sessionID)
	session := sessionIFace.(Session)

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
				if err := session.BroadcastError(uid, err.Error()); err != nil {
					break
				}

				if removeErr := session.RemoveConn(uid); removeErr != nil {
					break
				}

				break
			}

			switch msg.Type {
			case msgTypeRead:
				if err := session.RequestAccess(uid, msg); err != nil {
					break
				}
			default:
				if err := session.BroadcastToSet(msg, session.Conns[uid].Subscriptions); err != nil {
					break
				}
			}
		}
	}).ServeHTTP(c.Response(), c.Request())

	return nil
}
