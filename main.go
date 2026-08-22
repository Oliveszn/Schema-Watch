package main

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/Oliveszn/Schema-Watch/internal/alert"
	"github.com/Oliveszn/Schema-Watch/internal/config"
	"github.com/Oliveszn/Schema-Watch/internal/dashboard"
	"github.com/Oliveszn/Schema-Watch/internal/proxy"
	"github.com/Oliveszn/Schema-Watch/internal/store"
	"github.com/gin-gonic/gin"
)

func main() {
	target := flag.String("target", "", "backend URL to proxy requests to (overrides config.yaml)")
	port := flag.String("port", "", "port for schema-watch to listen on (overrides config.yaml)")
	configPath := flag.String("config", "config.yaml", "path to config file (optional)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("no config file found at %s, using flags/defaults only", *configPath)
			cfg = &config.Config{}
		} else {
			log.Fatalf("failed to load config: %v", err)
		}
	}

	targetURL := firstNonEmpty(*target, cfg.Target, "http://localhost:8080")
	listenPort := firstNonEmpty(*port, cfg.Port, "9090")

	st := store.New()
	console := alert.NewConsole()

	p, err := proxy.New(targetURL, st, console.Alert, cfg)
	if err != nil {
		log.Fatalf("failed to set up proxy: %v", err)
	}

	r := gin.Default()

	dashboard.New(st).RegisterRoutes(r)
	r.NoRoute(gin.WrapH(p.Handler()))

	log.Printf("schema-watch listening on :%s, forwarding to %s", listenPort, targetURL)
	log.Printf("point your frontend at http://localhost:%s instead of %s", listenPort, targetURL)
	log.Printf("dashboard: http://localhost:%s/__schema-watch/dashboard", listenPort)

	if err := r.Run(":" + listenPort); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
