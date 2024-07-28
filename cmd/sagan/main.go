package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kgaughan/sagan/internal/config"
	"github.com/kgaughan/sagan/internal/ui"
	"github.com/kgaughan/sagan/internal/version"
	flag "github.com/spf13/pflag"
)

func main() {
	flag.Parse()

	if *PrintVersion {
		fmt.Println(version.Version)
		return
	}
	if *ShowHelp {
		flag.Usage()
		os.Exit(0)
	}

	cfg := &config.Config{}
	if err := cfg.Load(*ConfigPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("%#v", cfg)

	p := tea.NewProgram(ui.InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
