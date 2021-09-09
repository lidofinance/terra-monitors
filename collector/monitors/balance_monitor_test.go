package monitors

import (
	"context"
	"io/ioutil"

	"github.com/stretchr/testify/suite"
)

type BalanceMonitorTestSuite struct {
	suite.Suite
}

func (suite *BalanceMonitorTestSuite) SetupTest() {

}

func (suite *BalanceMonitorTestSuite) TestBalance() {
	balanceInfo, err := ioutil.ReadFile("./test_data/balance_monitor_uusd_found.json")
	suite.NoError(err)

	ts := NewServerWithResponse(string(balanceInfo))
	cfg := NewTestCollectorConfig(ts.URL)

	logger := NewTestLogger()
	m := NewOperatorBotBalanceMonitor(cfg, logger)

	err = m.Handler(context.Background())
	suite.NoError(err)
	expectedBalance := 828.498499

	suite.InDelta(expectedBalance, m.balanceUST.Get(), 1e-6)
}

func (suite *BalanceMonitorTestSuite) TestNoUUSDBalance() {
	balanceInfo, err := ioutil.ReadFile("./test_data/balance_monitor_no_uusd_found.json")
	suite.NoError(err)

	ts := NewServerWithResponse(string(balanceInfo))
	cfg := NewTestCollectorConfig(ts.URL)

	logger := NewTestLogger()

	m := NewOperatorBotBalanceMonitor(cfg, logger)
	err = m.Handler(context.Background())
	suite.Error(err)

	expectedErr := "uusd coin not found"

	suite.EqualError(err, expectedErr)

	expectedBalance := 0
	suite.InDelta(expectedBalance, m.balanceUST.Get(), 1e-6)
}
