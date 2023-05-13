package main

import "flag"

var cfgPath = flag.String(
	"config",
	"sagan.yaml",
	"Path to configuration file",
)

var interactive = flag.Bool(
	"interactive",
	false,
	"Run in interactive mode",
)

var printVersion = flag.Bool(
	"version",
	false,
	"Print version and exit",
)
