package monitors

import (
	"context"
	"fmt"
	"testing"

	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/stretchr/testify/suite"
)

type MonitorTestSuite struct {
	suite.Suite
}

func (suite *MonitorTestSuite) SetupTest() {

}

func (suite *MonitorTestSuite) TestSuccessfulQueryRequest() {
	totalSupply := 79178685320809.0
	expected := types.TokenInfoResponse{
		Name:        "Bonded Luna",
		Symbol:      "BLUNA",
		Decimals:    6,
		TotalSupply: fmt.Sprintf("%.0f", totalSupply),
	}
	ts := NewServerWithResponse(BlunaTokenInfo)
	cfg := NewTestCollectorConfig(ts.URL)
	logger := NewTestLogger()
	blunaTokenInfoMonitor := NewBlunaTokenInfoMonitor(cfg, logger)

	err := blunaTokenInfoMonitor.Handler(context.Background())
	suite.Require().NoError(err)
	suite.Equal(expected, *blunaTokenInfoMonitor.State)
	suite.Equal(totalSupply, blunaTokenInfoMonitor.GetMetrics()[BlunaTotalSupply].Get())
}

func (suite *MonitorTestSuite) TestBadQueryRequest() {
	expectedErr := "bad query"
	ts := NewServerWithError(expectedErr)
	cfg := NewTestCollectorConfig(ts.URL)
	logger := NewTestLogger()
	blunaTokenInfoMonitor := NewBlunaTokenInfoMonitor(cfg, logger)

	err := blunaTokenInfoMonitor.Handler(context.Background())
	suite.Require().Error(err)
	suite.Contains(err.Error(), expectedErr)
}

func (suite *MonitorTestSuite) TestConnectionRefusedRequest() {
	expectedErr := "connection refused"
	ts := NewServerWithClosedConnectionError()
	cfg := NewTestCollectorConfig(ts.URL)
	logger := NewTestLogger()
	blunaTokenInfoMonitor := NewBlunaTokenInfoMonitor(cfg, logger)

	err := blunaTokenInfoMonitor.Handler(context.Background())
	suite.Require().Error(err)
	suite.Contains(err.Error(), expectedErr)
}

func TestLocales(t *testing.T) {
	suite.Run(t, new(MonitorTestSuite))
	suite.Run(t, new(UpdateGlobalIndexMonitorTestSuite))
	suite.Run(t, new(SlashingMonitorTestSuite))
	suite.Run(t, new(DetectorChangesTestSuite))
	suite.Run(t, new(ValidatorsCommissionTestSuite))
	suite.Run(t, new(ValidatorsRepositoryTestSuite))
	suite.Run(t, new(V2ValidatorsRepositoryTestSuite))
	suite.Run(t, new(OracleVotesMonitorTestSuite))
	suite.Run(t, new(BalanceMonitorTestSuite))
	suite.Run(t, new(FailedRedelegationsMonitorTestSuite))
}
