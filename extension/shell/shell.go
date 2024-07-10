package shell

import zero "github.com/KomeiDiSanXian/Koumakan"

func Parse(s string) []string {
	return zero.ParseShell(s)
}
