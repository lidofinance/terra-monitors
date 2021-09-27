package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/monitors/signinfo"
	"github.com/lidofinance/terra-monitors/openapi/models"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"strconv"
)

type MissedBlocksMonitorTestSuite struct {
	suite.Suite
}

func (suite *MissedBlocksMonitorTestSuite) SetupTest() {

}

func (suite *MissedBlocksMonitorTestSuite) TestValConsPub() {
	valcons := "terravalcons1qw4gg8v3jt0tfaq2qv337sa22slj3dxu73tyql"
	valconspub := "terravalconspub1zcjduepq5zcrunelz9yy09ksug5tcvx7r46mslnxk9gxqp8xflmwm3md8aesw6u3a8"

	expectedValConsAddr, err := ValConsToAddr(valcons)
	suite.NoError(err)
	actualValConsAddr, err := ValConsPubToAddr(valconspub)
	suite.NoError(err)
	suite.Equal(expectedValConsAddr, actualValConsAddr)
}

func (suite *MissedBlocksMonitorTestSuite) TestMissedBlocks() {
	suite.testMissedBlocks(config.NetworkGenerationColumbus5)
}

func (suite *MissedBlocksMonitorTestSuite) testMissedBlocks(networkGeneration string) {
	// validators's address is present in block_info's signatures
	// moniker - Test validator
	validatorInfoDataSignedBlock, err := ioutil.ReadFile(fmt.Sprintf("./test_data/%s/validators/first.json", networkGeneration))
	suite.NoError(err)

	// validators's address is not present in block_info's signatures
	// moniker - Test validator2
	validatorInfoDataNotSignedBlock, err := ioutil.ReadFile(fmt.Sprintf("./test_data/%s/validators/second.json", networkGeneration))
	suite.NoError(err)

	blockInfoBz, err := ioutil.ReadFile(fmt.Sprintf("./test_data/%s/block_info.json", networkGeneration))
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile(fmt.Sprintf("./test_data/%s/validators/two_whitelisted_validators.json", networkGeneration))
	suite.NoError(err)

	testServerResponses := map[string]string{
		fmt.Sprintf("/staking/validators/%s", testValAddress):  string(validatorInfoDataSignedBlock),
		fmt.Sprintf("/staking/validators/%s", testValAddress2): string(validatorInfoDataNotSignedBlock),
		"/blocks/latest": string(blockInfoBz),
		fmt.Sprintf("/wasm/contracts/%s/store", HubContract): string(whitelistedValidators),
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

	testServer := NewServerWithRoutedResponse(testServerResponses)

	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V1Contracts
	cfg.NetworkGeneration = networkGeneration

	logger := NewTestLogger()
	valRepository := NewValidatorsRepository(cfg, logger)
	signInfoRepository := signinfo.NewSignInfoRepository(cfg, logger)

	m := NewMissedBlocksMonitor(cfg, logger, valRepository, signInfoRepository)
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

	//there are some errors due to not implemented test endpoint /blocks/height
	fmt.Println(logger.Out)
}
