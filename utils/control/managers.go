package control

import (
	"sort"

	zero "github.com/wdvxdr1123/ZeroBot"
)

// ForEachByPrio iterates through managers by their priority.
func ForEachByPrio(iterator func(i int, manager IControl[*zero.Ctx]) bool) {
	for i, v := range cpmp2lstbyprio() {
		if !iterator(i, v) {
			return
		}
	}
}

func cpmp2lstbyprio() []IControl[*zero.Ctx] {
	managers.rw.RLock()
	defer managers.rw.RUnlock()
	ret := make([]IControl[*zero.Ctx], 0, len(managers.m))
	for _, v := range managers.m {
		ret = append(ret, v)
	}
	sort.SliceStable(ret, func(i, j int) bool {
		return enmap[ret[i].GetServiceName()].prio < enmap[ret[j].GetServiceName()].prio
	})
	return ret
}
