package main

import (
	"flag"
	"log"
	"os"

	"github.com/pha06ntom/lenta-scraper/internal/app"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "configs/example.yaml", "Path to yaml config")
	flag.Parse()

	a, err := app.NewFromYAML(cfgPath)
	if err != nil {
		log.Fatal("config: %v", err)
	}

	if err := a.Run(); err != nil {
		log.Println("error:", err)
		os.Exit(1)
	}
}
