package collector

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-openapi/runtime"
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
	fmt.Println(r)
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

type response struct {
	res *http.Response
}

func (r response) Code() int {
	return r.res.StatusCode
}

func (r response) Message() string {
	return r.res.Status
}

func (r response) GetHeader(name string) string {
	return r.res.Header.Get(name)
}

func (r response) Body() io.ReadCloser {
	return r.res.Body
}

type MockTransport struct {
	client   HttpClient
	consumer runtime.Consumer
}

func (t MockTransport) Submit(operation *runtime.ClientOperation) (interface{}, error) {
	fmt.Println(*operation)
	readResponse := operation.Reader

	req, err := http.NewRequest("GET", "https://lala", nil)
	if err != nil {
		return nil, err
	}
	// req.URL.Scheme = r.pickScheme(operation.Schemes)
	// req.URL.Host = r.Host
	// req.Host = r.Host

	var client HttpClient
	client = operation.Client
	if client == (*http.Client)(nil) {
		client = t.client
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	cons := t.consumer

	return readResponse.ReadResponse(response{res}, cons)
}

func (t *MockTransport) SetConsumer(consumer runtime.Consumer) {
	t.consumer = consumer
}

func NewMockTransport(client HttpClient) MockTransport {
	return MockTransport{
		client:   client,
		consumer: runtime.JSONConsumer(),
	}
}
