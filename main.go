package main

import (
	"flag"
	"net/http"

	"github.com/lidofinance/terra-monitors/app"
	"github.com/lidofinance/terra-monitors/collector"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/monitors"
	"github.com/lidofinance/terra-monitors/extractor"
)

var addr = flag.String("listen-address", ":8080",
	"The address to listen on for HTTP requests.")

func createCollector() collector.LCDCollector {
	defConfig := config.DefaultCollectorConfig()
	c := collector.NewLCDCollector(defConfig)
	hubStateMonitor := monitors.NewHubStateMonitor(defConfig)
	c.RegisterMonitor(&hubStateMonitor)

	rewardStateMonitor := monitors.NewRewardStateMonitor(defConfig)
	c.RegisterMonitor(&rewardStateMonitor)

	blunaTokenInfoMonitor := monitors.NewBlunaTokenInfoMonitor(defConfig)
	c.RegisterMonitor(&blunaTokenInfoMonitor)

	validatorsRepository := monitors.NewV1ValidatorsRepository(defConfig)
	slashingMonitor := monitors.NewSlashingMonitor(defConfig, validatorsRepository)
	c.RegisterMonitor(slashingMonitor)

	updateGlobalIndexMonitor := monitors.NewUpdateGlobalIndexMonitor(defConfig)
	c.RegisterMonitor(&updateGlobalIndexMonitor)

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
