package control

import (
	"sync"

	sql "github.com/FloatTech/sqlite"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
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

// IControlEngineTrigger is an interface for trigger.
type IControlEngineTrigger interface {
	On(typ string, rules ...zero.Rule) IControlMatcher
	OnCommand(commands string, rules ...zero.Rule) IControlMatcher
	OnCommandGroup(commands []string, rules ...zero.Rule) IControlMatcher
	OnFullMatch(src string, rules ...zero.Rule) IControlMatcher
	OnFullMatchGroup(src []string, rules ...zero.Rule) IControlMatcher
	OnKeyword(keyword string, rules ...zero.Rule) IControlMatcher
	OnKeywordGroup(keywords []string, rules ...zero.Rule) IControlMatcher
	OnMessage(rules ...zero.Rule) IControlMatcher
	OnMetaEvent(rules ...zero.Rule) IControlMatcher
	OnNotice(rules ...zero.Rule) IControlMatcher
	OnPrefix(prefix string, rules ...zero.Rule) IControlMatcher
	OnPrefixGroup(prefix []string, rules ...zero.Rule) IControlMatcher
	OnRegex(regexPattern string, rules ...zero.Rule) IControlMatcher
	OnRequest(rules ...zero.Rule) IControlMatcher
	OnShell(command string, model any, rules ...zero.Rule) IControlMatcher
	OnSuffix(suffix string, rules ...zero.Rule) IControlMatcher
	OnSuffixGroup(suffix []string, rules ...zero.Rule) IControlMatcher
}

// IControlEngineHandler is an interface for handler.
type IControlEngineHandler interface {
	UsePreHandler(rules ...zero.Rule)
	UsePostHandler(handler ...zero.Handler)
	UseMidHandler(rules ...zero.Rule)
}

// IControlEngine is an interface for ControlEngine.
type IControlEngine interface {
	IControlEngineTrigger
	IControlEngineHandler
	ApplySingle(s *single.Single[int64]) IControlEngine
	DataFolder() string
	Delete()
	GetCustomLazyData(dataurl, filename string) ([]byte, error)
	GetLazyData(filename string, isDataMustEqual bool) ([]byte, error)
	InitWhenNoError(errfun func() error, do func())
	IsEnabledIn(id int64) bool

	getPrio() int
}

type IControlMatcher interface {
	Handle(handler zero.Handler)
	Limit(limiterfn func(*zero.Ctx) *rate.Limiter, postfn ...func(*zero.Ctx)) IControlMatcher
	SetBlock(block bool) IControlMatcher
}
