package db

import (
	"context"

	"github.com/labstack/echo/v4"
)

// DBContext describes the basic echo context required
// by handlers.
type DBContext struct {
	echo.Context
	TLADB
}

// TLADB describes the basic set of operations
// required by backend handlers.
// Atomicity of operations on a TLADB are
// implementation-dependent.
type TLADB interface {
	LoadProgram(context.Context, string) (Program, error)
	StoreProgram(context.Context, Program) error

	LoadClass(context.Context, string) (Class, error)
	StoreClass(context.Context, Class) error
	DeleteClass(context.Context, Class) error

	LoadUser(context.Context, string) (User, error)
	StoreUser(context.Context, User) error
}
