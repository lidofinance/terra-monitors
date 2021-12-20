package stubs

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

	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories"
	"github.com/lidofinance/terra-monitors/internal/app/collector/types"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/logging"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
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
		Source: config.Source{
			Endpoints: endpoints,
			Schemes:   []string{"http"},
		},
		Addresses: config.Addresses{
			HubContract:                 types.HubContract,
			RewardContract:              types.RewardContract,
			BlunaTokenInfoContract:      types.BlunaTokenInfoContract,
			UpdateGlobalIndexBotAddress: types.UpdateGlobalIndexBotAddress,
			RewardsDispatcherContract:   types.RewardDispatcherContract,
			ValidatorsRegistryContract:  types.ValidatorsRegistryContract,
			AirDropRegistryContract:     types.AirDropRegistryContract,
		},
	}

	return cfg
}

func BuildValidatorsRepositoryConfig(cfg config.CollectorConfig) repositories.ValidatorsRepositoryConfig {
	return repositories.ValidatorsRepositoryConfig{
		BAssetContractsVersion:     cfg.BassetContractsVersion,
		HubContract:                cfg.Addresses.HubContract,
		ValidatorsRegistryContract: cfg.Addresses.ValidatorsRegistryContract,
	}
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
	rtr := mux.NewRouter()

	for route, response := range routeToResponse {
		responseValue := response

		// A bit ugly, but gets the job done.
		if strings.Contains(response, "error") {
			rtr.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(400)
				_, _ = fmt.Fprintf(w, `%s`, responseValue)

			})
		} else {
			rtr.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-Type", "application/json")
				_, _ = fmt.Fprintln(w, responseValue)
			})
		}
	}

	return httptest.NewServer(rtr)
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
	dir, err := utils.GetTerraMonitorsPath()
	if err != nil {
		panic(err)
	}

	resp := map[string]interface{}{
		"next":  offset - 10,
		"limit": 10,
		"txs":   []interface{}{},
	}
	txDataRaw, err := ioutil.ReadFile(fmt.Sprintf(dir+"test_data/%s/update_global_index_template.json", networkGeneration))
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
