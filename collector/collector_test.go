package collector

import (
	"bytes"
	"context"
	"testing"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/lidofinance/terra-monitors/internal/logging"
	"github.com/stretchr/testify/suite"
)

type CollectorTestSuite struct {
	suite.Suite
	Collector LCDCollector
}

func (suite *CollectorTestSuite) SetupTest() {
	logger := logging.NewDefaultLogger()
	out := bytes.NewBuffer(nil)
	logger.Out = out
	suite.Collector = NewLCDCollector(logger)
}

func (suite *CollectorTestSuite) TestSuccessfullQueryRequest() {
	expected := TokenInfoResponse{
		Name:        "Bonded Luna",
		Symbol:      "BLUNA",
		Decimals:    6,
		TotalSupply: "79178685320809",
	}
	ts := NewServerWithResponse(BlunaTokenInfo)
	cfg := NewTransportConfig(ts.URL)
	transport := httptransport.New(cfg.Host, cfg.BasePath, cfg.Schemes)
	suite.Collector.SetTransport(transport)
	blunaTokenInfoMonitor := NewBlunaTokenInfoMintor("bluna_token_contract_address")
	suite.Collector.RegisterMonitor(&blunaTokenInfoMonitor)
	err := blunaTokenInfoMonitor.Handler(context.Background())
	suite.Require().NoError(err)
	suite.Equal(expected, *blunaTokenInfoMonitor.State)
}

func (suite *CollectorTestSuite) TestBadQueryRequest() {
	expectedErr := "bad query"
	ts := NewServerWithError(expectedErr)
	cfg := NewTransportConfig(ts.URL)
	transport := httptransport.New(cfg.Host, cfg.BasePath, cfg.Schemes)
	suite.Collector.SetTransport(transport)
	blunaTokenInfoMonitor := NewBlunaTokenInfoMintor("bluna_token_contract_address")
	suite.Collector.RegisterMonitor(&blunaTokenInfoMonitor)
	err := blunaTokenInfoMonitor.Handler(context.Background())
	suite.Require().Error(err)
	suite.Contains(err.Error(), expectedErr)
}

func (suite *CollectorTestSuite) TestConnectionRefusedRequest() {
	expectedErr := "connection refused"
	ts := NewServerWithClosedConnectionError()
	cfg := NewTransportConfig(ts.URL)
	transport := httptransport.New(cfg.Host, cfg.BasePath, cfg.Schemes)
	suite.Collector.SetTransport(transport)
	blunaTokenInfoMonitor := NewBlunaTokenInfoMintor("bluna_token_contract_address")
	suite.Collector.RegisterMonitor(&blunaTokenInfoMonitor)
	err := blunaTokenInfoMonitor.Handler(context.Background())
	suite.Require().Error(err)
	suite.Contains(err.Error(), expectedErr)
}

func TestLocales(t *testing.T) {
	suite.Run(t, new(CollectorTestSuite))
}
