package monitors

import (
	"bytes"
	"context"
	"testing"

	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/lidofinance/terra-monitors/internal/logging"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type MonitorTestSuite struct {
	suite.Suite
	logger *logrus.Logger
}

func (suite *MonitorTestSuite) SetupTest() {
	suite.logger = logging.NewDefaultLogger()
	out := bytes.NewBuffer(nil)
	suite.logger.Out = out
}

func (suite *MonitorTestSuite) TestSuccessfullQueryRequest() {
	expected := types.TokenInfoResponse{
		Name:        "Bonded Luna",
		Symbol:      "BLUNA",
		Decimals:    6,
		TotalSupply: "79178685320809",
	}
	ts := NewServerWithResponse(BlunaTokenInfo)
	cfg := NewTransportConfig(ts.URL)
	apiClient := client.NewHTTPClientWithConfig(nil, cfg)
	blunaTokenInfoMonitor := NewBlunaTokenInfoMintor("bluna_token_contract_address", apiClient, nil)
	blunaTokenInfoMonitor.SetApiClient(apiClient)
	blunaTokenInfoMonitor.SetLogger(suite.logger)

	err := blunaTokenInfoMonitor.Handler(context.Background())
	suite.Require().NoError(err)
	suite.Equal(expected, *blunaTokenInfoMonitor.State)
}

func (suite *MonitorTestSuite) TestBadQueryRequest() {
	expectedErr := "bad query"
	ts := NewServerWithError(expectedErr)
	cfg := NewTransportConfig(ts.URL)
	apiClient := client.NewHTTPClientWithConfig(nil, cfg)
	blunaTokenInfoMonitor := NewBlunaTokenInfoMintor("bluna_token_contract_address", apiClient, nil)
	blunaTokenInfoMonitor.SetApiClient(apiClient)
	blunaTokenInfoMonitor.SetLogger(suite.logger)

	err := blunaTokenInfoMonitor.Handler(context.Background())
	suite.Require().Error(err)
	suite.Contains(err.Error(), expectedErr)
}

func (suite *MonitorTestSuite) TestConnectionRefusedRequest() {
	expectedErr := "connection refused"
	ts := NewServerWithClosedConnectionError()
	cfg := NewTransportConfig(ts.URL)
	apiClient := client.NewHTTPClientWithConfig(nil, cfg)
	blunaTokenInfoMonitor := NewBlunaTokenInfoMintor("bluna_token_contract_address", apiClient, nil)
	blunaTokenInfoMonitor.SetApiClient(apiClient)
	blunaTokenInfoMonitor.SetLogger(suite.logger)

	err := blunaTokenInfoMonitor.Handler(context.Background())
	suite.Require().Error(err)
	suite.Contains(err.Error(), expectedErr)
}

func TestLocales(t *testing.T) {
	suite.Run(t, new(MonitorTestSuite))
}
