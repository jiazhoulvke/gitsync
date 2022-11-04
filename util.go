package main

import (
	"os"
	"path/filepath"
)

func AbsPath(p string) string {
	if p == "" {
		return ""
	}
	if p[0] == '~' {
		p = "$HOME" + p[1:]
	}
	p = os.ExpandEnv(p)
	p, _ = filepath.Abs(p)
	return p
}
