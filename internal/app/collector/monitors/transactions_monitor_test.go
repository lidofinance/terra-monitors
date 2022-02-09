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

func (suite *TransactionsMonitorTestSuite) TestTransactionsHeight() {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	// FIX
	transactionsInfo, err := ioutil.ReadFile(dir + "test_data/transactions_monitor_found.json")
	suite.NoError(err)

	ts := stubs.NewServerWithResponse(string(transactionsInfo))
	cfg := stubs.NewTestCollectorConfig(ts.URL)

	logger := stubs.NewTestLogger()
	m := NewTransactionsMonitor(cfg, logger)

	err = m.Handler(context.Background())
	suite.NoError(err)
	// FIX
	expectedHeight := 234234

	// FIX
	suite.InDelta(expectedHeight, m.lastTransactionHeight.Get(), 1e-6)
}
