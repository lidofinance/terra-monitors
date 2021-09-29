package monitors

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/stretchr/testify/suite"
)

type ValidatorsRepositoryTestSuite struct {
	suite.Suite
}

func (suite *ValidatorsRepositoryTestSuite) SetupTest() {

}

func (suite *ValidatorsRepositoryTestSuite) TestSuccessfulRequest() {
	validatorInfoData, err := ioutil.ReadFile("./test_data/columbus-5/slashing_validator_info_not_jailed.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile("./test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", testValAddress): string(validatorInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", HubContract):  string(whitelistedValidators),
	})
	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V1Contracts
	cfg.NetworkGeneration = config.NetworkGenerationColumbus5

	valRepository := NewValidatorsRepository(cfg, NewTestLogger())

	expectedValidators := []string{testValAddress}
	validators, err := valRepository.GetValidatorsAddresses(context.Background())
	suite.NoError(err)
	suite.Equal(validators, expectedValidators)

	expectedValidatorInfo := types.ValidatorInfo{
		Address:        testValAddress,
		PubKey:         testConsAddress,
		Moniker:        TestMoniker,
		CommissionRate: TestCommissionRate,
	}
	validatorInfo, err := valRepository.GetValidatorInfo(context.Background(), testValAddress)
	suite.NoError(err)
	suite.Equal(validatorInfo, expectedValidatorInfo)
}

func (suite *ValidatorsCommissionTestSuite) TestFailedValidatorsRepository() {
	validatorInfoData, err := ioutil.ReadFile("./test_data/slashing_error.json")
	suite.NoError(err)

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", testValAddress): string(validatorInfoData),

		// error format is the same
		fmt.Sprintf("/wasm/contracts/%s/store", HubContract): string(validatorInfoData),
	})
	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V1Contracts
	cfg.NetworkGeneration = config.NetworkGenerationColumbus5

	valRepository := NewValidatorsRepository(cfg, NewTestLogger())

	validators, err := valRepository.GetValidatorsAddresses(context.Background())
	suite.Nil(validators)
	suite.Error(err)

	_, err = valRepository.GetValidatorInfo(context.Background(), testValAddress)
	suite.Error(err)
}
