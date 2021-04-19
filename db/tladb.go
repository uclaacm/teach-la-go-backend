package db

import "context"

// TLADB describes the basic set of operations
// required by backend handlers.
type TLADB interface {
	LoadProgram(context.Context, string) (Program, error)
	StoreProgram(context.Context, *Program) error

	LoadClass(context.Context, string) (Class, error)
	StoreClass(context.Context, *Class) error

	LoadUser(context.Context, string) (User, error)
	StoreUser(context.Context, *User) error
}
