package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/lidofinance/terra-monitors/app"
	"github.com/lidofinance/terra-monitors/collector"
	"github.com/lidofinance/terra-monitors/collector/monitors"
	"github.com/lidofinance/terra-monitors/extractor"
	"github.com/lidofinance/terra-monitors/internal/logging"
	"github.com/sirupsen/logrus"
)

var (
	HubContract                  = "terra1mtwph2juhj0rvjz7dy92gvl6xvukaxu8rfv8ts"
	RewardContract               = "terra17yap3mhph35pcwvhza38c2lkj7gzywzy05h7l0"
	BlunaTokenInfoContract       = "terra1kc87mu460fwkqte29rquh4hc20m54fxwtsx7gp"
	UpdateGlobalIndexBotContract = "terra1eqpx4zr2vm9jwu2vas5rh6704f6zzglsayf2fy"
)

var addr = flag.String("listen-address", ":8080",
	"The address to listen on for HTTP requests.")

func createCollector(logger *logrus.Logger) collector.LCDCollector {
	c := collector.NewLCDCollector(
		logger,
	)
	hubStateMonitor := monitors.NewHubStateMintor(HubContract, c.GetApiClient(), logger)
	c.RegisterMonitor(&hubStateMonitor)

	rewardStateMonitor := monitors.NewRewardStateMintor(RewardContract, c.GetApiClient(), logger)
	c.RegisterMonitor(&rewardStateMonitor)

	blunaTokenInfoMonitor := monitors.NewBlunaTokenInfoMintor(BlunaTokenInfoContract, c.GetApiClient(), logger)
	c.RegisterMonitor(&blunaTokenInfoMonitor)

	updateBlobalIndexMonitor := monitors.NewUpdateGlobalIndexMonitor(UpdateGlobalIndexBotContract, c.GetApiClient(), logger)
	c.RegisterMonitor((&updateBlobalIndexMonitor))
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

func main2() {
	logger := logging.NewDefaultLogger()
	c := collector.NewLCDCollector(
		logger,
	)
	m := monitors.NewUpdateGlobalIndexMonitor(HubContract, c.GetApiClient(), nil)
	err := m.Handler(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	for _, l := range m.ApiResponse.Txs[0].Logs {
		for _, e := range l.Events {
			data, err := json.Marshal(e)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(string(data))
		}
	}

}
