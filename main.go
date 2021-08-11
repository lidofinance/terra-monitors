package main

import (
	"context"
	"flag"
	"github.com/lidofinance/terra-monitors/app"
	"github.com/lidofinance/terra-monitors/collector"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/monitors"
	"github.com/lidofinance/terra-monitors/extractor"
	"net/http"
)

var addr = flag.String("listen-address", ":8080",
	"The address to listen on for HTTP requests.")

func createCollector() collector.LCDCollector {
	ctx := context.Background()
	defConfig := config.DefaultCollectorConfig()
	c := collector.NewLCDCollector(defConfig)

	hubStateMonitor := monitors.NewHubStateMonitor(defConfig)
	c.RegisterMonitor(ctx, &hubStateMonitor)

	rewardStateMonitor := monitors.NewRewardStateMonitor(defConfig)
	c.RegisterMonitor(ctx, &rewardStateMonitor)

	blunaTokenInfoMonitor := monitors.NewBlunaTokenInfoMonitor(defConfig)
	c.RegisterMonitor(ctx, &blunaTokenInfoMonitor)

	validatorsRepository := monitors.NewV1ValidatorsRepository(defConfig)
	slashingMonitor := monitors.NewSlashingMonitor(defConfig, validatorsRepository)
	c.RegisterMonitor(ctx, slashingMonitor)

	updateGlobalIndexMonitor := monitors.NewUpdateGlobalIndexMonitor(defConfig)
	c.RegisterMonitor(ctx, updateGlobalIndexMonitor)

	hubParameters := monitors.NewHubParametersMonitor(defConfig)
	c.RegisterMonitor(ctx, &hubParameters)

	configCRC32Monitor := monitors.NewConfigsCRC32Monitor(defConfig)
	c.RegisterMonitor(ctx, &configCRC32Monitor)

	whitelistedValidatorsMonitor := monitors.NewWhitelistedValidatorsMonitor(defConfig, validatorsRepository)
	c.RegisterMonitor(ctx, &whitelistedValidatorsMonitor)

	validatorsFeeMonitor := monitors.NewValidatorsFeeMonitor(defConfig, validatorsRepository)
	c.RegisterMonitor(&validatorsFeeMonitor)

	return c
}

func main() {
	flag.Parse()
	c := createCollector()
	logger := c.GetLogger()
	p := extractor.NewPromExtractor(&c, logger)
	app := app.NewAppHTTP(p)
	http.Handle("/metrics", app)
	logger.Printf("Starting web server at %s\n", *addr)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		logger.Errorf("http.ListenAndServer: %v\n", err)
	}
}
