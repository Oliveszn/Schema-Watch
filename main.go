package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Oliveszn/Schema-Watch/internal/proxy"
	"github.com/Oliveszn/Schema-Watch/internal/schema"
	"github.com/Oliveszn/Schema-Watch/internal/store"
	"github.com/gin-gonic/gin"
)

func main() {
	target := flag.String("target", "http://localhost:8080", "backend URL to proxy requests to")
	port := flag.String("port", "9090", "port for schema-watch to listen on")
	flag.Parse()

	st := store.New()

	onDiff := func(d *schema.Diff) {
		printDiff(d)
	}

	p, err := proxy.New(*target, st, onDiff)
	if err != nil {
		log.Fatalf("failed to set up proxy: %v", err)
	}

	r := gin.Default()
	r.NoRoute(gin.WrapH(p.Handler()))

	log.Printf("schema-watch listening on :%s, forwarding to %s", *port, *target)
	log.Printf("point your frontend at http://localhost:%s instead of %s", *port, *target)

	if err := r.Run(":" + *port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func printDiff(d *schema.Diff) {
	label := "CHANGE"
	if d.Breaking {
		label = "BREAKING CHANGE"
	}
	fmt.Printf("\n[schema-watch] %s on %s\n", label, d.Endpoint)
	for _, c := range d.Changes {
		switch c.Type {
		case schema.FieldAdded:
			fmt.Printf("  + %s added (%s)\n", c.Path, c.NewType)
		case schema.FieldRemoved:
			fmt.Printf("  - %s removed (was %s)\n", c.Path, c.OldType)
		case schema.FieldTypeChanged:
			fmt.Printf("  ~ %s changed type: %s -> %s\n", c.Path, c.OldType, c.NewType)
		}
	}
}
