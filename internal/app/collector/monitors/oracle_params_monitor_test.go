package monitors

import (
	"context"
	"github.com/lidofinance/terra-monitors/internal/pkg/stubs"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
)

type OracleParamsMonitorTestSuite struct {
	suite.Suite
}

func (suite *OracleParamsMonitorTestSuite) SetupTest() {

}

func (suite *OracleParamsMonitorTestSuite) TestSlashingParamsSuccessful() {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	respData, err := ioutil.ReadFile(dir + "test_data/oracle_params.json")
	suite.NoError(err)

	ts := stubs.NewServerWithResponse(string(respData))
	cfg := stubs.NewTestCollectorConfig(ts.URL)

	logger := stubs.NewTestLogger()
	oracleParamsMonitor := NewOracleParamsMonitor(cfg, logger)

	err = oracleParamsMonitor.Handler(context.Background())
	suite.NoError(err)

	blocksWindowExpected := 432_000.0
	suite.Equal(blocksWindowExpected, oracleParamsMonitor.metrics[OracleMissedVotesWindow].Get())
}
