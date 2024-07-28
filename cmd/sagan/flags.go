package main

import (
	"fmt"
	"os"
	"path"

	"github.com/kgaughan/sagan/internal/version"
	flag "github.com/spf13/pflag"
)

var (
	ConfigPath   = flag.StringP("config", "c", "./sagan.yaml", "path to configuration file")
	PrintVersion = flag.BoolP("version", "V", false, "print version and exit")
	ShowHelp     = flag.BoolP("help", "h", false, "show help")
)

func init() {
	flag.Usage = func() {
		name := path.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "%s (v%s) - a task runner\n\n", name, version.Version)
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
}
