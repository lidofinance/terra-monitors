package monitors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"

	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/internal/logging"
	"github.com/sirupsen/logrus"
)

const BlunaTokenInfo = `{"height":"3754668","result":{"name":"Bonded Luna","symbol":"BLUNA","decimals":6,"total_supply":"79178685320809"}}`

const (
	HubContract                 = "terra1mtwph2juhj0rvjz7dy92gvl6xvukaxu8rfv8ts"
	RewardContract              = "terra17yap3mhph35pcwvhza38c2lkj7gzywzy05h7l0"
	BlunaTokenInfoContract      = "terra1kc87mu460fwkqte29rquh4hc20m54fxwtsx7gp"
	UpdateGlobalIndexBotAddress = "dummy_updateglobalindexbot"
	RewardDispatcherContract    = "dummy_rewarddispatcher"
	ValidatorsRegistryContract  = "dummy_validatorsregistry"
	AirDropRegistryContract     = "dummy_airdropRegistry"
)

func NewTestCollectorConfig(urlsWithScheme ...string) config.CollectorConfig {
	var endpoints []string
	for _, urlWithScheme := range urlsWithScheme {
		parsedURL, err := url.Parse(urlWithScheme)
		if err != nil {
			log.Fatalf("failed to parse URL %s: %s\n", urlWithScheme, err)
		}

		endpoints = append(endpoints, parsedURL.Host)
	}

	cfg := config.CollectorConfig{
		LCD: config.LCD{
			Endpoints: endpoints,
			Schemes:   []string{"http"},
		},
		Addresses: config.Addresses{
			HubContract:                 HubContract,
			RewardContract:              RewardContract,
			BlunaTokenInfoContract:      BlunaTokenInfoContract,
			UpdateGlobalIndexBotAddress: UpdateGlobalIndexBotAddress,
			RewardsDispatcherContract:   RewardDispatcherContract,
			ValidatorsRegistryContract:  ValidatorsRegistryContract,
			AirDropRegistryContract:     AirDropRegistryContract,
		},
	}

	return cfg
}

func NewTestLogger() *logrus.Logger {
	logger := logging.NewDefaultLogger()
	out := bytes.NewBuffer(nil)
	logger.Out = out

	return logger
}

func NewServerWithResponse(resp string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, resp)
	}))

	return ts
}

func NewServerWithRoutedResponse(routeToResponse map[string]string) *httptest.Server {
	mux := http.NewServeMux()

	for route, response := range routeToResponse {
		responseValue := response

		// A bit ugly, but gets the job done.
		if strings.Contains(response, "error") {
			mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(400)
				_, _ = fmt.Fprintf(w, `%s`, responseValue)

			})
		} else {
			mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-Type", "application/json")
				_, _ = fmt.Fprintln(w, responseValue)
			})
		}
	}

	return httptest.NewServer(mux)
}

func NewServerWithRandomJson() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		response := make(map[string]interface{})
		response["height"] = "10000"
		response["result"] = map[string]int{
			"data": rand.Int(),
		}
		data, err := json.Marshal(response)
		if err != nil {
			panic(err)
		}
		_, _ = fmt.Fprintln(w, string(data))
	}))
	return ts
}

func NewServerWithError(errorMessage string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		_, _ = fmt.Fprintf(w, `{"error":"%s"}`, errorMessage)
	}))
	return ts
}

func NewServerWithClosedConnectionError() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.Close()
	return ts
}

func makeTxFromData(data []byte) map[string]interface{} {
	txData := make(map[string]interface{})
	err := json.Unmarshal(data, &txData)
	if err != nil {
		panic(err)
	}
	return txData
}

func makeTxs(networkGeneration string, offset int) string {
	resp := map[string]interface{}{
		"next":  offset - 10,
		"limit": 10,
		"txs":   []interface{}{},
	}
	txDataRaw, err := ioutil.ReadFile(fmt.Sprintf("./test_data/%s/update_global_index_template.json", networkGeneration))
	if err != nil {
		panic(err)
	}
	var txData map[string]interface{}
	for i := 0; i < 10; i++ {
		txData = makeTxFromData(txDataRaw)
		txData["id"] = offset - i
		resp["txs"] = append(resp["txs"].([]interface{}), txData)
	}
	respRaw, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	return string(respRaw)
}

func NewServerForUpdateGlobalIndex(networkGeneration string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offsetRaw := r.URL.Query().Get("offset")
		if offsetRaw == "" {
			offsetRaw = "200"
		}
		offset, err := strconv.Atoi(offsetRaw)
		if err != nil {
			panic(err)
		}
		w.Header().Add("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, makeTxs(networkGeneration, offset))
	}))
	return ts
}
