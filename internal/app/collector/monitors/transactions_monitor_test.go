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

	expectedHeight := 6201254
	testAddress := cfg.MonitoredAccountAddresses[0]
	actualHeight := m.metricVectors[LastTransactionHeight].Get(testAddress)
	suite.InDelta(expectedHeight, actualHeight, 1e-6)
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