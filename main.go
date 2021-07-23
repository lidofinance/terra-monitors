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
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/monitors"
	"github.com/lidofinance/terra-monitors/extractor"
	"github.com/lidofinance/terra-monitors/internal/logging"
	"github.com/sirupsen/logrus"
)

var addr = flag.String("listen-address", ":8080",
	"The address to listen on for HTTP requests.")

func createCollector(logger *logrus.Logger) collector.LCDCollector {
	defConfig := config.DefaultCollectorConfig()
	c := collector.NewLCDCollector(defConfig)
	hubStateMonitor := monitors.NewHubStateMintor(defConfig)
	c.RegisterMonitor(&hubStateMonitor)

	rewardStateMonitor := monitors.NewRewardStateMonitor(defConfig)
	c.RegisterMonitor(&rewardStateMonitor)

	blunaTokenInfoMonitor := monitors.NewBlunaTokenInfoMintor(defConfig)
	c.RegisterMonitor(&blunaTokenInfoMonitor)

	updateBlobalIndexMonitor := monitors.NewUpdateGlobalIndexMonitor(defConfig)
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
	defConfig := config.DefaultCollectorConfig()

	// c := collector.NewLCDCollector(
	// 	defConfig,
	// )
	m := monitors.NewUpdateGlobalIndexMonitor(defConfig)
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
