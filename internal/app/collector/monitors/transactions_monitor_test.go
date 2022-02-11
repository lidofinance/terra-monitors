package monitors

import (
	"context"
	"io/ioutil"

	"github.com/lidofinance/terra-monitors/internal/pkg/stubs"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"
	"github.com/stretchr/testify/suite"
)

type TransactionsMonitorTestSuite struct {
	suite.Suite
}

func (suite *TransactionsMonitorTestSuite) SetupTest() {
}

func (suite *TransactionsMonitorTestSuite) TestTransactionsHeightFound() {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	transactionsInfo, err := ioutil.ReadFile(dir + "test_data/transactions_monitor_found.json")
	suite.NoError(err)

	ts := stubs.NewServerWithResponse(string(transactionsInfo))
	cfg := stubs.NewTestCollectorConfig(ts.URL)

	logger := stubs.NewTestLogger()
	m := NewTransactionsMonitor(cfg, logger)

	err = m.Handler(context.Background())
	suite.NoError(err)

	expectedID := 214871418
	testAddress := cfg.Addresses.MonitoredAccounts[0]
	actualID := m.metricVectors[MonitoredTransactionId].Get(testAddress)
	suite.InDelta(expectedID, actualID, 1e-6)
}

func (suite *TransactionsMonitorTestSuite) TestTransactionsHeightNotFound() {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	transactionsInfo, err := ioutil.ReadFile(dir + "test_data/transactions_monitor_not_found.json")
	suite.NoError(err)

	ts := stubs.NewServerWithResponse(string(transactionsInfo))
	cfg := stubs.NewTestCollectorConfig(ts.URL)

	logger := stubs.NewTestLogger()
	m := NewTransactionsMonitor(cfg, logger)

	err = m.Handler(context.Background())
	suite.Error(err)
}
