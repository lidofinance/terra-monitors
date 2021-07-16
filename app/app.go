package app

import (
	"net/http"

	"github.com/lidofinance/terra-monitors/extractor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type AppHTTP struct {
	prom extractor.PromExtractor
}

func (a AppHTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.prom.UpdateMetrics(r.Context())
	promhttp.Handler().ServeHTTP(w, r)
}

func NewAppHTTP(p extractor.PromExtractor) AppHTTP {
	return AppHTTP{
		prom: p,
	}
}
