package util

import (
	"flag"
)

type flags struct {
	ConfPath string
}

var Flags flags

func ParseFlags() {
	flag.StringVar(&Flags.ConfPath, "config", "", "path to config JSON file")
	flag.Parse()
	return
}
