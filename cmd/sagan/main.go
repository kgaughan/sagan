package main

import (
	"fmt"
	"log"

	"github.com/kgaughan/sagan/internal/config"
)

func main() {
	cfg := &config.Config{}
	if err := cfg.Load("test.yaml"); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%#v", cfg)
}
