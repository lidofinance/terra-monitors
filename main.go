package main

import (
	"flag"
	"net/http"

	"github.com/lidofinance/terra-monitors/app"
	"github.com/lidofinance/terra-monitors/collector"
	"github.com/lidofinance/terra-monitors/extractor"
	"github.com/lidofinance/terra-monitors/internal/logging"
	"github.com/sirupsen/logrus"
)

var (
	HubContract            = "terra1mtwph2juhj0rvjz7dy92gvl6xvukaxu8rfv8ts"
	RewardContract         = "terra17yap3mhph35pcwvhza38c2lkj7gzywzy05h7l0"
	BlunaTokenInfoContract = "terra1kc87mu460fwkqte29rquh4hc20m54fxwtsx7gp"
)

var addr = flag.String("listen-address", ":8080",
	"The address to listen on for HTTP requests.")

func createCollector(logger *logrus.Logger) collector.LCDCollector {
	c := collector.NewLCDCollector(
		logger,
	)
	hubStateMonitor := collector.NewHubStateMintor(HubContract)
	c.RegisterMonitor(&hubStateMonitor)

	rewardStateMonitor := collector.NewRewardStateMintor(RewardContract)
	c.RegisterMonitor(&rewardStateMonitor)

	blunaTokenInfoMonitor := collector.NewBlunaTokenInfoMintor(BlunaTokenInfoContract)
	c.RegisterMonitor(&blunaTokenInfoMonitor)
	return c
}

func main() {
	flag.Parse()
	logger := logging.NewDefaultLogger()
	c := createCollector(logger)
	p := extractor.NewPromExtractor(&c, logger)
	app := app.NewAppHTTP(p)
	http.Handle("/metrics", app)

	logger.Printf("Starting web server at %s\n", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		logger.Errorf("http.ListenAndServer: %v\n", err)
	}
}
