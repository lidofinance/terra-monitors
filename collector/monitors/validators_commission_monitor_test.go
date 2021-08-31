package monitors

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/lidofinance/terra-monitors/collector/config"

	"github.com/stretchr/testify/suite"
)

const (
	TestMoniker        = "Test validator"
	TestCommissionRate = 0.08
)

type ValidatorsCommissionTestSuite struct {
	suite.Suite
}

func (suite *ValidatorsCommissionTestSuite) SetupTest() {

}

func (suite *ValidatorsCommissionTestSuite) TestSuccessfulRequest() {
	validatorInfoData, err := ioutil.ReadFile("./test_data/slashing_validator_info_not_jailed.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile("./test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", testValAddress): string(validatorInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", HubContract):  string(whitelistedValidators),
	})
	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V1Contracts

	valRepository := NewValidatorsRepository(cfg)
	logger := NewTestLogger()
	m := NewValidatorsFeeMonitor(cfg, logger, valRepository)
	err = m.Handler(context.Background())
	suite.NoError(err)

	metricVectors := m.GetMetricVectors()

	expectedValidatorsCommission := 0.08
	actualValidatorsCommission := metricVectors[ValidatorsCommission].Get(TestMoniker)

	suite.Equal(expectedValidatorsCommission, actualValidatorsCommission)
}

func (suite *ValidatorsCommissionTestSuite) TestFailedValidatorsFeeRequest() {
	validatorInfoData, err := ioutil.ReadFile("./test_data/slashing_error.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile("./test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", testValAddress): string(validatorInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", HubContract):  string(whitelistedValidators),
	})
	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = config.V1Contracts

	valRepository := NewValidatorsRepository(cfg)
	logger := NewTestLogger()
	m := NewValidatorsFeeMonitor(cfg, logger, valRepository)
	err = m.Handler(context.Background())
	suite.Error(err)
}
