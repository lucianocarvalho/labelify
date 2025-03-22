package main

import (
	"log"
	"net/http"

	"github.com/lucianocarvalho/labelify/internal/config"
	"github.com/lucianocarvalho/labelify/internal/infrastructure"
	"github.com/lucianocarvalho/labelify/internal/usecase"
)

func main() {
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Error loading config.yaml: %v", err)
	}

	labelifyConfig, err := config.LoadLabelifyConfig(cfg.ConfigPath)
	if err != nil {
		log.Fatalf("Error loading Labelify config: %v", err)
	}

	hydrate := usecase.NewHydrateUseCase(labelifyConfig)

	proxy, err := infrastructure.NewProxy(cfg.PrometheusURL, hydrate)
	if err != nil {
		log.Fatalf("Error creating proxy: %v", err)
	}

	proxy.SetupModifyResponse()

	http.Handle("/", proxy)

	log.Printf("Proxy listening on http://localhost:%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
