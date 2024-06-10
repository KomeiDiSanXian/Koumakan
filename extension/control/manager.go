package control

import (
	"encoding/json"
	"errors"
	"math/bits"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/sqlite"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

var (
	blockCache = make(map[int64]bool)
	respCache  = make(map[int64]string)
)

// Manager is the manager of Control
type Manager[Ctx any] struct {
	sync.RWMutex
	Controls map[string]*Control[Ctx]
	DataBase sql.Sqlite
}

// NewManager creates a new Manager
func NewManager[Ctx any](dbpath string) (m Manager[Ctx]) {
	switch {
	case dbpath == "":
		dbpath = "ctrl.db"
	case strings.HasSuffix(dbpath, "/"):
		err := os.MkdirAll(dbpath, 0755)
		if err != nil {
			panic(err)
		}
		dbpath += "ctrl.db"
	default:
		i := strings.LastIndex(dbpath, "/")
		if i > 0 {
			err := os.MkdirAll(dbpath[:i], 0755)
			if err != nil {
				panic(err)
			}
		}
	}
	m = Manager[Ctx]{
		Controls: make(map[string]*Control[Ctx]),
		DataBase: sql.Sqlite{DBPath: dbpath},
	}
	err := m.DataBase.Open(time.Hour) // cache 1 hour
	if err != nil {
		panic(err)
	}
	err = m.initBlock()
	if err != nil {
		panic(err)
	}
	err = m.initResponse()
	if err != nil {
		panic(err)
	}
	return
}

// initBlock initializes the block table
func (m *Manager[Ctx]) initBlock() error {
	return m.DataBase.Create("__block", &BlockStatus{})
}

// initResponse initializes the response table
func (manager *Manager[CTX]) initResponse() error {
	return manager.DataBase.Create("__resp", &ResponseGroup{})
}

// GetControl gets the control by service
func (m *Manager[Ctx]) GetControl(service string) (*Control[Ctx], bool) {
	m.RLock()
	defer m.RUnlock()
	ctrl, ok := m.Controls[service]
	return ctrl, ok
}

// ForEach iterates all controls
func (m *Manager[Ctx]) ForEach(f func(k string, v *Control[Ctx]) bool) {
	m.RLock()
	ctrl := copyControl(m.Controls)
	m.RUnlock()
	for k, v := range ctrl {
		if !f(k, v) {
			return
		}
	}
}

// copyControl copies the control map
func copyControl[Ctx any](m map[string]*Control[Ctx]) map[string]*Control[Ctx] {
	n := make(map[string]*Control[Ctx])
	for k, v := range m {
		n[k] = v
	}
	return n
}

// NewControl creates a new control with options
func (m *Manager[Ctx]) NewControl(service string, opt *Option[Ctx]) *Control[Ctx] {
	ctrl := &Control[Ctx]{
		Service: service,
		Cache:   make(map[int64]uint8),
		Options: func() Option[Ctx] {
			if opt != nil {
				return *opt
			}
			return Option[Ctx]{}
		}(),
		Manager: m,
	}
	m.Lock()
	defer m.Unlock()
	m.Controls[service] = ctrl
	var gconf GroupConfig
	if err := m.DataBase.Create(service, &gconf); err != nil {
		panic(err)
	}
	if err := m.DataBase.Create(service+"ban", &BanStatus{}); err != nil {
		panic(err)
	}
	if err := m.DataBase.Find(ctrl.Service, &gconf, "WHERE gid = 0"); err == nil {
		if bits.RotateLeft64(uint64(gconf.Disable), 1)&1 == 1 {
			ctrl.Options.DefaultDisable = !ctrl.Options.DefaultDisable
		}
	}
	return ctrl
}

// DoBlock blocks user
func (m *Manager[Ctx]) DoBlock(uid int64) error {
	m.Lock()
	defer m.Unlock()
	blockCache[uid] = true
	return m.DataBase.Insert("__block", &BlockStatus{UserID: uid})
}

// DoUnblock unblocks user
func (m *Manager[Ctx]) DoUnblock(uid int64) error {
	m.Lock()
	defer m.Unlock()
	delete(blockCache, uid)
	return m.DataBase.Del("__block", "WHERE uid = "+strconv.FormatInt(uid, 10))
}

// IsBlocked checks if user is blocked
func (m *Manager[Ctx]) IsBlocked(uid int64) bool {
	m.RLock()
	isBlocked, ok := blockCache[uid]
	m.RUnlock()
	if ok {
		return isBlocked
	}
	m.Lock()
	defer m.Unlock()
	isBlocked = m.DataBase.CanFind("__block", "WHERE uid = "+strconv.FormatInt(uid, 10))
	blockCache[uid] = isBlocked
	return isBlocked
}

// CanResponse checks if the group can response
func (m *Manager[Ctx]) CanResponse(gid int64) bool {
	m.RLock()
	ext, ok := respCache[0] // all status
	m.RUnlock()
	if ok && ext != "-" {
		return true
	}
	m.RLock()
	ext, ok = respCache[gid]
	m.RUnlock()
	if ok {
		return ext != "-"
	}
	m.RLock()
	var rsp ResponseGroup
	err := m.DataBase.Find("__resp", &rsp, "where gid = 0") // all status
	m.RUnlock()
	if err == nil && rsp.Extra != "-" {
		m.Lock()
		respCache[0] = rsp.Extra
		m.Unlock()
		return true
	}
	m.RLock()
	err = m.DataBase.Find("__resp", &rsp, "where gid = "+strconv.FormatInt(gid, 10))
	m.RUnlock()
	if err != nil {
		m.Lock()
		respCache[gid] = "-"
		m.Unlock()
		return false
	}
	m.Lock()
	respCache[gid] = rsp.Extra
	m.Unlock()
	return rsp.Extra != "-"
}

// Response opens response for group
func (m *Manager[Ctx]) Response(gid int64) error {
	if m.CanResponse(gid) {
		return errors.New("group " + strconv.FormatInt(gid, 10) + " already has response")
	}
	m.Lock()
	respCache[gid] = ""
	err := m.DataBase.Insert("__resp", &ResponseGroup{GroupID: gid})
	m.Unlock()
	return err
}

// Silence drops extra for group
func (m *Manager[Ctx]) Silence(gid int64) error {
	if !m.CanResponse(gid) {
		return errors.New("group " + strconv.FormatInt(gid, 10) + " already has silence")
	}
	m.Lock()
	respCache[gid] = "-"
	err := m.DataBase.Del("__resp", "where gid = "+strconv.FormatInt(gid, 10))
	m.Unlock()
	return err
}

func (m *Manager[Ctx]) getExtra(groupID int64, obj any) error {
	if !m.CanResponse(groupID) {
		return errors.New("group " + strconv.FormatInt(groupID, 10) + " has silence, no extra")
	}
	m.RLock()
	ext, ok := respCache[groupID]
	m.RUnlock()
	if ok {
		if ext == "-" {
			return errors.New("group " + strconv.FormatInt(groupID, 10) + " has empty extra")
		}
		return json.Unmarshal(helper.StringToBytes(ext), obj)
	}
	var rsp ResponseGroup
	m.RLock()
	err := m.DataBase.Find("__resp", &rsp, "where gid = "+strconv.FormatInt(groupID, 10))
	m.RUnlock()
	if err != nil || rsp.Extra == "-" {
		m.Lock()
		respCache[groupID] = "-"
		m.Unlock()
		return errors.New("group " + strconv.FormatInt(groupID, 10) + " has empty extra")
	}
	m.Lock()
	respCache[groupID] = rsp.Extra
	m.Unlock()
	return json.Unmarshal(helper.StringToBytes(rsp.Extra), obj)
}

func (m *Manager[Ctx]) setExtra(groupID int64, obj any) error {
	if !m.CanResponse(groupID) {
		return errors.New("group " + strconv.FormatInt(groupID, 10) + " has silence, no extra")
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	m.Lock()
	defer m.Unlock()
	respCache[groupID] = helper.BytesToString(data)
	return m.DataBase.Insert("__resp", &ResponseGroup{GroupID: groupID, Extra: helper.BytesToString(data)})
}
