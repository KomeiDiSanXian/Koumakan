package control

import (
	"sync"

	sql "github.com/FloatTech/sqlite"
)

// IManager is an interface for Manager.
type IManager[CTX any] interface {
	RW() *sync.RWMutex
	ControlMap() map[string]IControl[CTX]
	DB() *sql.Sqlite

	CanResponse(gid int64) bool
	DoBlock(uid int64) error
	DoUnblock(uid int64) error
	ForEach(iterator func(key string, manager IControl[CTX]) bool)
	IsBlocked(uid int64) bool
	Lookup(service string) (IControl[CTX], bool)
	NewControl(service string, options *Options[CTX]) IControl[CTX]
	Response(gid int64) error
	Silence(gid int64) error

	getExtra(gid int64, obj any) error
	initBlock() error
	initResponse() error
	setExtra(gid int64, obj any) error
}

// IControl is an interface for Control.
type IControl[CTX any] interface {
	Ban(uid int64, gid int64)
	Disable(groupID int64)
	Enable(groupID int64)
	EnableMarkIn(grp int64) EnableMark
	Flip() error
	GetData(gid int64) int64
	GetExtra(obj any) error
	Handler(gid int64, uid int64) bool
	IsBannedIn(uid int64, gid int64) bool
	IsEnabledIn(gid int64) bool
	Permit(uid int64, gid int64)
	Reset(groupID int64)
	SetData(groupID int64, data int64) error
	SetExtra(obj any) error
	String() string

	GetOptions() Options[CTX]
	GetServiceName() string
}
