package validators

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/lidofinance/terra-monitors/internal/app/collector/types"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/stubs"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"
	"github.com/stretchr/testify/suite"
)

type V2ValidatorsRepositoryTestSuite struct {
	suite.Suite
}

func (suite *V2ValidatorsRepositoryTestSuite) SetupTest() {

}

func (suite *V2ValidatorsRepositoryTestSuite) TestSuccessfulRequest() {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	validatorInfoData, err := ioutil.ReadFile(dir + "test_data/columbus-5/slashing_validator_info_not_jailed.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile(dir + "test_data/validators_registry_validators_response.json")
	suite.NoError(err)

	testServer := stubs.NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", types.TestValAddress):               string(validatorInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", types.ValidatorsRegistryContract): string(whitelistedValidators),
	})
	cfg := stubs.NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V2Contracts
	cfg.NetworkGeneration = config.NetworkGenerationColumbus5

	valRepository := NewValidatorsRepository(cfg, stubs.NewTestLogger())

	expectedValidatorsAddresses := []string{types.TestValAddress}
	validators, err := valRepository.GetValidatorsAddresses(context.Background())
	suite.NoError(err)
	suite.Equal(expectedValidatorsAddresses, validators)

	expectedValidatorInfo := types.ValidatorInfo{
		Address:        types.TestValAddress,
		PubKey:         types.TestConsAddress,
		Moniker:        types.TestMoniker,
		CommissionRate: types.TestCommissionRate,
	}
	validatorInfo, err := valRepository.GetValidatorInfo(context.Background(), types.TestValAddress)
	suite.NoError(err)
	suite.Equal(expectedValidatorInfo, validatorInfo)
}
