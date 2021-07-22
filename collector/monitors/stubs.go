package monitors

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/lidofinance/terra-monitors/openapi/client"
)

var BlunaTokenInfo = `{"height":"3754668","result":{"name":"Bonded Luna","symbol":"BLUNA","decimals":6,"total_supply":"79178685320809"}}`

var BadQuery = "{\"error\":\"contract query failed: parsing anchor_basset_hub::msg::QueryMsg: unknown variant `config1`, expected one of `config`, `state`, `whitelisted_validators`, `current_batch`, `withdrawable_unbonded`, `parameters`, `unbond_requests`, `all_history`\"}"

func NewTransportConfig(urlWithScheme string) *client.TransportConfig {
	host := strings.Split(urlWithScheme, "//")[1]
	return &client.TransportConfig{
		Host:     host,
		BasePath: "/",
		Schemes:  []string{"http"},
	}
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
