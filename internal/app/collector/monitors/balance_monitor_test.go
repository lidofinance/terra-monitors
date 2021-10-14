package monitors

import (
	"context"
	"io/ioutil"

	"github.com/lidofinance/terra-monitors/internal/pkg/stubs"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"
	"github.com/stretchr/testify/suite"
)

type BalanceMonitorTestSuite struct {
	suite.Suite
}

func (suite *BalanceMonitorTestSuite) SetupTest() {

}

func (suite *BalanceMonitorTestSuite) TestBalance() {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	balanceInfo, err := ioutil.ReadFile(dir + "test_data/balance_monitor_uusd_found.json")
	suite.NoError(err)

	ts := stubs.NewServerWithResponse(string(balanceInfo))
	cfg := stubs.NewTestCollectorConfig(ts.URL)

	logger := stubs.NewTestLogger()
	m := NewOperatorBotBalanceMonitor(cfg, logger)

	err = m.Handler(context.Background())
	suite.NoError(err)
	expectedBalance := 828.498499

	suite.InDelta(expectedBalance, m.balanceUST.Get(), 1e-6)
}

func (suite *BalanceMonitorTestSuite) TestNoUUSDBalance() {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	balanceInfo, err := ioutil.ReadFile(dir + "test_data/balance_monitor_no_uusd_found.json")
	suite.NoError(err)

	ts := stubs.NewServerWithResponse(string(balanceInfo))
	cfg := stubs.NewTestCollectorConfig(ts.URL)

	logger := stubs.NewTestLogger()

	m := NewOperatorBotBalanceMonitor(cfg, logger)
	err = m.Handler(context.Background())
	suite.Error(err)

	expectedErr := "uusd coin not found"

	suite.EqualError(err, expectedErr)

	expectedBalance := 0
	suite.InDelta(expectedBalance, m.balanceUST.Get(), 1e-6)
}
