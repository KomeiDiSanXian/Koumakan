// Package control 控制插件的启用与优先级等
package control

import (
	"os"
	"strings"
	"sync"
	"time"

	sql "github.com/FloatTech/sqlite"
)

// Manager 管理
type Manager[CTX any] struct {
	rw *sync.RWMutex
	m  map[string]IControl[CTX]
	d  *sql.Sqlite
}

// RW 返回读写锁
func (m *Manager[CTX]) RW() *sync.RWMutex {
	return m.rw
}

// ControlMap 返回控制器映射
func (m *Manager[CTX]) ControlMap() map[string]IControl[CTX] {
	return m.m
}

// DB 返回数据库
func (m *Manager[CTX]) DB() *sql.Sqlite {
	return m.d
}

// NewManager 打开管理数据库
func NewManager[CTX any](dbpath string) (m Manager[CTX]) {
	switch {
	case dbpath == "":
		dbpath = "ctrl.db"
	case strings.HasSuffix(dbpath, "/"):
		err := os.MkdirAll(dbpath, 0o755)
		if err != nil {
			panic(err)
		}
		dbpath += "ctrl.db"
	default:
		i := strings.LastIndex(dbpath, "/")
		if i > 0 {
			err := os.MkdirAll(dbpath[:i], 0o755)
			if err != nil {
				panic(err)
			}
		}
	}
	m = Manager[CTX]{
		m: map[string]IControl[CTX]{},
		d: &sql.Sqlite{DBPath: dbpath},
	}
	err := m.d.Open(time.Hour)
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

// Lookup returns a Manager by the service name, if
// not exist, it will return nil.
func (manager *Manager[CTX]) Lookup(service string) (IControl[CTX], bool) {
	manager.rw.RLock()
	m, ok := manager.m[service]
	manager.rw.RUnlock()
	return m, ok
}

// ForEach iterates through managers.
func (manager *Manager[CTX]) ForEach(iterator func(key string, manager IControl[CTX]) bool) {
	manager.rw.RLock()
	m := cpmp(manager.m)
	manager.rw.RUnlock()
	for k, v := range m {
		if !iterator(k, v) {
			return
		}
	}
}

func cpmp[CTX any](m map[string]IControl[CTX]) map[string]IControl[CTX] {
	ret := make(map[string]IControl[CTX], len(m))
	for k, v := range m {
		ret[k] = v
	}
	return ret
}
