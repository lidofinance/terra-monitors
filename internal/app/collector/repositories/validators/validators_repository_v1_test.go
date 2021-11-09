package validators

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/lidofinance/terra-monitors/internal/app/collector/types"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/stubs"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"
	"github.com/stretchr/testify/suite"
)

func TestValidatorsRepo(t *testing.T) {
	suite.Run(t, new(ValidatorsRepositoryTestSuite))
	suite.Run(t, new(V2ValidatorsRepositoryTestSuite))
}

type ValidatorsRepositoryTestSuite struct {
	suite.Suite
}

func (suite *ValidatorsRepositoryTestSuite) SetupTest() {

}

func (suite *ValidatorsRepositoryTestSuite) TestSuccessfulRequest() {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	validatorInfoData, err := ioutil.ReadFile(dir + "test_data/columbus-5/slashing_validator_info_not_jailed.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile(dir + "test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	testServer := stubs.NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", types.TestValAddress): string(validatorInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", types.HubContract):  string(whitelistedValidators),
	})
	cfg := stubs.NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V1Contracts
	cfg.NetworkGeneration = config.NetworkGenerationColumbus5

	valRepository := NewValidatorsRepository(cfg, stubs.NewTestLogger())

	expectedValidators := []string{types.TestValAddress}
	validators, err := valRepository.GetValidatorsAddresses(context.Background())
	suite.NoError(err)
	suite.Equal(validators, expectedValidators)

	expectedValidatorInfo := types.ValidatorInfo{
		Address:        types.TestValAddress,
		PubKey:         types.TestConsAddress,
		Moniker:        types.TestMoniker,
		CommissionRate: types.TestCommissionRate,
	}
	validatorInfo, err := valRepository.GetValidatorInfo(context.Background(), types.TestValAddress)
	suite.NoError(err)
	suite.Equal(validatorInfo, expectedValidatorInfo)
}
