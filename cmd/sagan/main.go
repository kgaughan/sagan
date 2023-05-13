package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/kgaughan/sagan/internal/config"
	"github.com/kgaughan/sagan/internal/version"
)

func main() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		name := path.Base(os.Args[0])
		fmt.Fprintf(out, "%s (v%s) - a task runner\n\n", name, version.Version)
		fmt.Fprintf(out, "Usage:\n  %s [flags]\n\n", name)
		fmt.Fprintf(out, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *printVersion {
		fmt.Println(version.Version)
		return
	}
	cfg := &config.Config{}
	if err := cfg.Load(*cfgPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%v %#v", *interactive, cfg)
}
