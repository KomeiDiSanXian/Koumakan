package zero

import (
	"sync/atomic"

	"github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/extension/control"
)

var (
	enmap       = make(map[string]Engine)
	prio        uint64
	custpriomap map[string]uint64
)

// LoadCustomPriority 加载自定义优先级
func LoadCustomPriority(m map[string]uint64) {
	if custpriomap != nil {
		panic("double-defined custpriomap")
	}
	custpriomap = m
	prio = uint64(len(custpriomap)+1) * 10
}

// Register 注册引擎
func Register(service string, o *control.Option[*Ctx]) Engine {
	if custpriomap != nil {
		logrus.Debugf("[ZeroBot] Registering %s with custom priority", service)
		engine := newEngine(service, int(custpriomap[service]), o)
		enmap[service] = engine
		return engine
	}
	logrus.Debugf("[ZeroBot] Registering %s with default priority %d", service, prio)
	engine := newEngine(service, int(atomic.AddUint64(&prio, 10)), o) // prio += 10
	enmap[service] = engine
	return engine
}

// Delete 删除插件控制器, 数据不会被删除
func Delete(service string) {
	engine, ok := enmap[service]
	if ok {
		engine.Delete()
		managers.RLock()
		_, ok := managers.Controls[service]
		managers.RUnlock()
		if ok {
			managers.Lock()
			delete(managers.Controls, service)
			managers.Unlock()
		}
	}
}
