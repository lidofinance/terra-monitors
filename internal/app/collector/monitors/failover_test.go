package monitors

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/lidofinance/terra-monitors/internal/pkg/stubs"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"
	"github.com/stretchr/testify/suite"
)

type FailoverTestSuite struct {
	suite.Suite
}

func (suite *FailoverTestSuite) SetupTest() {

}

func (suite *FailoverTestSuite) TestFailover() {
	// N.B.: this test reuses an existing Balance test.
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	balanceInfo, err := ioutil.ReadFile(dir + "test_data/balance_monitor_uusd_found.json")
	suite.NoError(err)

	// Try to run the test with an incorrect URL (an error is expected).
	logger := stubs.NewTestLogger()
	incorrectURL := "http://127.0.0.1:1234"
	cfg := stubs.NewTestCollectorConfig(incorrectURL)
	m := NewOperatorBotBalanceMonitor(cfg, logger)
	err = m.Handler(context.Background())
	suite.Error(err)

	// Try to run the test with an incorrect URL prioritized over the correct one (no error is expected,
	// an error log is expected).
	logger = stubs.NewTestLogger()
	connectionRefusedLogMessagePattern := "connect: connection refused"
	ts := stubs.NewServerWithResponse(string(balanceInfo))
	cfg = stubs.NewTestCollectorConfig(incorrectURL, ts.URL)
	m = NewOperatorBotBalanceMonitor(cfg, logger)
	err = m.Handler(context.Background())
	suite.NoError(err)
	actualMessages := fmt.Sprintln(logger.Out)
	suite.Contains(actualMessages, connectionRefusedLogMessagePattern)

	// Try to run the test with a correct URL prioritized over the incorrect one (no error is expected,
	// no error log is expected).
	logger = stubs.NewTestLogger()
	cfg = stubs.NewTestCollectorConfig(ts.URL, incorrectURL)
	m = NewOperatorBotBalanceMonitor(cfg, logger)
	err = m.Handler(context.Background())
	suite.NoError(err)
	actualMessages = fmt.Sprintln(logger.Out)
	suite.NotContains(actualMessages, connectionRefusedLogMessagePattern)
}
