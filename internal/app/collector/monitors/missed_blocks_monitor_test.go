package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories/validators"
	"github.com/lidofinance/terra-monitors/internal/app/collector/types"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/stubs"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"
	"github.com/lidofinance/terra-monitors/openapi/models"
	"github.com/stretchr/testify/suite"
)

type MissedBlocksMonitorTestSuite struct {
	suite.Suite
}

func (suite *MissedBlocksMonitorTestSuite) SetupTest() {

}

func (suite *MissedBlocksMonitorTestSuite) TestMissedBlocks() {
	suite.testMissedBlocks(config.NetworkGenerationColumbus5)
}

func (suite *MissedBlocksMonitorTestSuite) testMissedBlocks(networkGeneration string) {
	// validators's address is present in block_info's signatures
	// moniker - Test validator
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	validatorInfoDataSignedBlock, err := ioutil.ReadFile(fmt.Sprintf(dir+"test_data/%s/validators/first.json", networkGeneration))
	suite.NoError(err)

	// validators's address is not present in block_info's signatures
	// moniker - Test validator2
	validatorInfoDataNotSignedBlock, err := ioutil.ReadFile(fmt.Sprintf(dir+"test_data/%s/validators/second.json", networkGeneration))
	suite.NoError(err)

	blockInfoBz, err := ioutil.ReadFile(fmt.Sprintf(dir+"test_data/%s/block_info.json", networkGeneration))
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile(fmt.Sprintf(dir+"test_data/%s/validators/two_whitelisted_validators.json", networkGeneration))
	suite.NoError(err)

	testServerResponses := map[string]string{
		fmt.Sprintf("/staking/validators/%s", types.TestValAddress):  string(validatorInfoDataSignedBlock),
		fmt.Sprintf("/staking/validators/%s", types.TestValAddress2): string(validatorInfoDataNotSignedBlock),
		"/blocks/latest": string(blockInfoBz),
		fmt.Sprintf("/wasm/contracts/%s/store", types.HubContract): string(whitelistedValidators),
	}
	blockInfo := models.BlockQuery{}
	err = json.Unmarshal(blockInfoBz, &blockInfo)
	suite.NoError(err)
	for i := 2; i <= 11; i++ {
		blockInfo.Block.LastCommit.Height = strconv.Itoa(i)
		blockInfoUpdated, err := json.Marshal(blockInfo)
		suite.NoError(err)
		testServerResponses[fmt.Sprintf("/blocks/%d", i)] = string(blockInfoUpdated)
	}

	testServer := stubs.NewServerWithRoutedResponse(testServerResponses)

	cfg := stubs.NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V1Contracts
	cfg.NetworkGeneration = networkGeneration

	logger := stubs.NewTestLogger()
	valRepository := validators.NewValidatorsRepository(cfg, logger)

	m := NewMissedBlocksMonitor(cfg, logger, valRepository)
	err = m.Handler(context.Background())
	suite.NoError(err)

	metricVectors := m.GetMetricVectors()

	expectedValidatorsLabelsCount := 2
	suite.Equal(expectedValidatorsLabelsCount, len(metricVectors[MissedBlocksForPeriod].Labels()))
	//"Test validator" has signed the block
	suite.Equal(0.0, metricVectors[MissedBlocksForPeriod].Get("Test validator"))
	// "Test validator2" has not signed the block
	// we have checked 10 blocks and  all 11 with no "Test validators2" sign
	suite.Equal(10.0, metricVectors[MissedBlocksForPeriod].Get("Test validator2"))

}
