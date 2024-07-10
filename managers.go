package zero

import (
	"sort"

	"github.com/KomeiDiSanXian/Koumakan/extension/control"
)

// ForEachByPrio iterates through managers by their priority.
func ForEachByPrio(iterator func(i int, manager *control.Control[*Ctx]) bool) {
	for i, v := range cpmp2lstbyprio() {
		if !iterator(i, v) {
			return
		}
	}
}

func cpmp2lstbyprio() []*control.Control[*Ctx] {
	managers.RLock()
	defer managers.RUnlock()
	ret := make([]*control.Control[*Ctx], 0, len(managers.Controls))
	for _, v := range managers.Controls {
		ret = append(ret, v)
	}
	sort.SliceStable(ret, func(i, j int) bool {
		return enmap[ret[i].Service].(*ZeroEngine).prio < enmap[ret[j].Service].(*ZeroEngine).prio
	})
	return ret
}
