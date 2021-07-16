package collector

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
)

var BlunaTokenInfo = `{"height":"3754668","result":{"name":"Bonded Luna","symbol":"BLUNA","decimals":6,"total_supply":"79178685320809"}}`

var BadQuery = "{\"error\":\"contract query failed: parsing anchor_basset_hub::msg::QueryMsg: unknown variant `config1`, expected one of `config`, `state`, `whitelisted_validators`, `current_batch`, `withdrawable_unbonded`, `parameters`, `unbond_requests`, `all_history`\"}"

type MockHttpClient struct {
	BlunaTokenInfo string
	BadQuery       string
}

func NewMockClient() MockHttpClient {
	return MockHttpClient{
		BlunaTokenInfo: BlunaTokenInfo,
		BadQuery:       BadQuery,
	}
}

func (c MockHttpClient) Do(r *http.Request) (*http.Response, error) {
	query := r.URL.Query().Get("query_msg")
	var resp *http.Response
	var err error
	if strings.Contains(query, "token_info") {
		resp = &http.Response{
			Body: io.NopCloser(bytes.NewBufferString(c.BlunaTokenInfo)),
		}
	} else if strings.Contains(query, "bad_query") {
		resp = &http.Response{
			Body: io.NopCloser(bytes.NewBufferString(c.BadQuery)),
		}
	} else {
		err = errors.New("connection refused")
	}
	return resp, err
}
