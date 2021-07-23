package monitors

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/internal/logging"
)

const BlunaTokenInfo = `{"height":"3754668","result":{"name":"Bonded Luna","symbol":"BLUNA","decimals":6,"total_supply":"79178685320809"}}`

const BadQuery = "{\"error\":\"contract query failed: parsing anchor_basset_hub::msg::QueryMsg: unknown variant `config1`, expected one of `config`, `state`, `whitelisted_validators`, `current_batch`, `withdrawable_unbonded`, `parameters`, `unbond_requests`, `all_history`\"}"

func NewTestCollectorConfig(urlWithScheme string) config.CollectorConfig {
	host := strings.Split(urlWithScheme, "//")[1]
	out := bytes.NewBuffer(nil)
	cfg := config.CollectorConfig{
		LCDEndpoint:            host,
		Logger:                 logging.NewDefaultLogger(),
		HubContract:            "terra1mtwph2juhj0rvjz7dy92gvl6xvukaxu8rfv8ts",
		RewardContract:         "terra17yap3mhph35pcwvhza38c2lkj7gzywzy05h7l0",
		BlunaTokenInfoContract: "terra1kc87mu460fwkqte29rquh4hc20m54fxwtsx7gp",
		Schemes:                []string{"http"},
	}
	cfg.Logger.Out = out
	return cfg
}

func NewServerWithResponse(resp string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintln(w, resp)
	}))
	return ts
}

func NewServerWithError(errorMessage string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		fmt.Fprintf(w, `{"error":"%s"}`, errorMessage)
	}))
	return ts
}

func NewServerWithClosedConnectionError() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.Close()
	return ts
}
