package monitors

import (
	"context"
	"github.com/lidofinance/terra-monitors/internal/pkg/stubs"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
)

type SlashingParamsMonitorTestSuite struct {
	suite.Suite
}

func (suite *SlashingParamsMonitorTestSuite) SetupTest() {

}

func (suite *SlashingParamsMonitorTestSuite) TestSlashingParamsSuccessful() {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	respData, err := ioutil.ReadFile(dir + "test_data/slashing_params_data.json")
	suite.NoError(err)

	ts := stubs.NewServerWithResponse(string(respData))
	cfg := stubs.NewTestCollectorConfig(ts.URL)

	logger := stubs.NewTestLogger()
	slashingParamsMonitor := NewSlashingParamsMonitor(cfg, logger)

	err = slashingParamsMonitor.Handler(context.Background())
	suite.NoError(err)

	blocksWindowExpected := 10_000.0
	suite.Equal(blocksWindowExpected, slashingParamsMonitor.metrics[SlashingSignedBlocksWindow].Get())
}
