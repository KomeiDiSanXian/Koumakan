package control

import (
	"errors"
	"strconv"
)

// InitResponse ...
func (manager *Manager[CTX]) initResponse() error {
	return manager.d.Create("__resp", &ResponseGroup{})
}

var respCache = make(map[int64]string)

// Response opens the resp of the gid
func (manager *Manager[CTX]) Response(gid int64) error {
	if manager.CanResponse(gid) {
		return errors.New("group " + strconv.FormatInt(gid, 10) + " already in response")
	}
	manager.rw.Lock()
	respCache[gid] = ""
	err := manager.d.Insert("__resp", &ResponseGroup{GroupID: gid})
	manager.rw.Unlock()
	return err
}

// Silence will drop its extra data
func (manager *Manager[CTX]) Silence(gid int64) error {
	if !manager.CanResponse(gid) {
		return errors.New("group " + strconv.FormatInt(gid, 10) + " already in silence")
	}
	manager.rw.Lock()
	respCache[gid] = "-"
	err := manager.d.Del("__resp", "where gid = "+strconv.FormatInt(gid, 10))
	manager.rw.Unlock()
	return err
}

// CanResponse ...
func (manager *Manager[CTX]) CanResponse(gid int64) bool {
	manager.rw.RLock()
	ext, ok := respCache[0] // all status
	manager.rw.RUnlock()
	if ok && ext != "-" {
		return true
	}
	manager.rw.RLock()
	ext, ok = respCache[gid]
	manager.rw.RUnlock()
	if ok {
		return ext != "-"
	}
	manager.rw.RLock()
	var rsp ResponseGroup
	err := manager.d.Find("__resp", &rsp, "where gid = 0") // all status
	manager.rw.RUnlock()
	if err == nil && rsp.Extra != "-" {
		manager.rw.Lock()
		respCache[0] = rsp.Extra
		manager.rw.Unlock()
		return true
	}
	manager.rw.RLock()
	err = manager.d.Find("__resp", &rsp, "where gid = "+strconv.FormatInt(gid, 10))
	manager.rw.RUnlock()
	if err != nil {
		manager.rw.Lock()
		respCache[gid] = "-"
		manager.rw.Unlock()
		return false
	}
	manager.rw.Lock()
	respCache[gid] = rsp.Extra
	manager.rw.Unlock()
	return rsp.Extra != "-"
}
