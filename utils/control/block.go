package control

import (
	"strconv"
)

func (manager *Manager[CTX]) initBlock() error {
	return manager.d.Create("__block", &BlockStatus{})
}

var blockCache = make(map[int64]bool)

// DoBlock 封禁
func (manager *Manager[CTX]) DoBlock(uid int64) error {
	manager.rw.Lock()
	defer manager.rw.Unlock()
	blockCache[uid] = true
	return manager.d.Insert("__block", &BlockStatus{UserID: uid})
}

// DoUnblock 解封
func (manager *Manager[CTX]) DoUnblock(uid int64) error {
	manager.rw.Lock()
	defer manager.rw.Unlock()
	blockCache[uid] = false
	return manager.d.Del("__block", "where uid = "+strconv.FormatInt(uid, 10))
}

// IsBlocked 是否封禁
func (manager *Manager[CTX]) IsBlocked(uid int64) bool {
	manager.rw.RLock()
	isbl, ok := blockCache[uid]
	manager.rw.RUnlock()
	if ok {
		return isbl
	}
	manager.rw.Lock()
	defer manager.rw.Unlock()
	isbl = manager.d.CanFind("__block", "where uid = "+strconv.FormatInt(uid, 10))
	blockCache[uid] = isbl
	return isbl
}