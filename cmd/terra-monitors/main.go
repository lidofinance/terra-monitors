package main

import (
	"flag"
	"net/http"

	"github.com/lidofinance/terra-monitors/internal/app"
	"github.com/lidofinance/terra-monitors/internal/app/collector"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/app/extractor"
	"github.com/lidofinance/terra-monitors/internal/pkg/logging"
)

var addr = flag.String("listen-address", ":8080",
	"The address to listen on for HTTP requests.")

func main() {
	flag.Parse()

	logger := logging.NewDefaultLogger()

	cfg, err := config.NewCollectorConfig()
	if err != nil {
		logger.Fatalf("Failed to create NewCollectorConfig: %s", err)
	}

	col, err := collector.New(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to create collector: %s", err)
	}

	var (
		promExtractor = extractor.NewPromExtractor(col, logger)
		appInstance   = app.NewAppHTTP(promExtractor)
	)
	http.Handle("/metrics", appInstance)
	logger.Printf("Starting web server v%s at %s\n", cfg.BassetContractsVersion, *addr)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		logger.Errorf("Failed to ListenAndServe: %v\n", err)
	}
}
