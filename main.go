package main

import (
	"context"
	"flag"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/lidofinance/terra-monitors/internal/logging"

	"github.com/lidofinance/terra-monitors/app"
	"github.com/lidofinance/terra-monitors/collector"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/monitors"
	"github.com/lidofinance/terra-monitors/extractor"
)

var addr = flag.String("listen-address", ":8080",
	"The address to listen on for HTTP requests.")

func createCollector(cfg config.CollectorConfig, logger *logrus.Logger) (collector.LCDCollector, error) {
	ctx := context.Background()

	c := collector.NewLCDCollector(cfg, logger)

	hubStateMonitor := monitors.NewHubStateMonitor(cfg, logger)
	c.RegisterMonitor(cfg, ctx, &hubStateMonitor)

	rewardStateMonitor := monitors.NewRewardStateMonitor(cfg, logger)
	c.RegisterMonitor(cfg, ctx, &rewardStateMonitor)

	blunaTokenInfoMonitor := monitors.NewBlunaTokenInfoMonitor(cfg, logger)
	c.RegisterMonitor(cfg, ctx, blunaTokenInfoMonitor)

	validatorsRepository := monitors.NewV1ValidatorsRepository(cfg)
	slashingMonitor := monitors.NewSlashingMonitor(cfg, logger, validatorsRepository)
	c.RegisterMonitor(cfg, ctx, slashingMonitor)

	updateGlobalIndexMonitor := monitors.NewUpdateGlobalIndexMonitor(cfg, logger)
	c.RegisterMonitor(cfg, ctx, updateGlobalIndexMonitor)

	hubParameters := monitors.NewHubParametersMonitor(cfg, logger)
	c.RegisterMonitor(cfg, ctx, &hubParameters)

	configCRC32Monitor := monitors.NewConfigsCRC32Monitor(cfg, logger)
	c.RegisterMonitor(cfg, ctx, &configCRC32Monitor)

	whitelistedValidatorsMonitor := monitors.NewWhitelistedValidatorsMonitor(cfg, logger, validatorsRepository)
	c.RegisterMonitor(cfg, ctx, &whitelistedValidatorsMonitor)

	validatorsFeeMonitor := monitors.NewValidatorsFeeMonitor(cfg, logger, validatorsRepository)
	c.RegisterMonitor(cfg, ctx, validatorsFeeMonitor)

	return c, nil
}

func main() {
	flag.Parse()

	logger := logging.NewDefaultLogger()

	cfg, err := config.NewCollectorConfig()
	if err != nil {
		logger.Fatalf("Failed to create NewCollectorConfig: %s", err)
	}

	col, err := createCollector(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to createCollector(): %s", err)
	}

	var (
		promExtractor = extractor.NewPromExtractor(&col, logger)
		appInstance   = app.NewAppHTTP(promExtractor)
	)
	http.Handle("/metrics", appInstance)

	logger.Printf("Starting web server at %s\n", *addr)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		logger.Errorf("Failed to ListenAndServe: %v\n", err)
	}
}
