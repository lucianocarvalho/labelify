package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/lucianocarvalho/labelify/internal/config"
	"github.com/lucianocarvalho/labelify/internal/infrastructure"
	"github.com/lucianocarvalho/labelify/internal/usecase"
)

func main() {
	flags := config.ParseFlags()

	cfg, err := config.LoadLabelifyConfig(flags.ConfigFile)
	if err != nil {
		log.Fatalf("Error loading config from %s: %v", flags.ConfigFile, err)
	}

	enrichmentUseCase := usecase.NewEnrichmentUseCase(cfg)

	proxy, err := infrastructure.NewProxy(cfg.Config.Prometheus.URL, enrichmentUseCase)
	if err != nil {
		log.Fatalf("Error creating proxy: %v", err)
	}

	http.Handle("/", proxy)

	log.Printf("Proxy listening on http://localhost:%d", cfg.Config.Server.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Config.Server.Port), nil))
}
