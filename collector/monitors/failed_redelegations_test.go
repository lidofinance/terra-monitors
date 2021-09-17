package monitors

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/stretchr/testify/suite"
)

const TestFailedRedelegationValidatorAddress = "terravaloper1qxqrtvg3smlfdfhvwcdzh0huh4f50kfs6gdt4x"

type FailedRedelegationsMonitorTestSuite struct {
	suite.Suite
}

func (suite *FailedRedelegationsMonitorTestSuite) SetupTest() {

}

func (suite *FailedRedelegationsMonitorTestSuite) TestRedelegationFailedRequest() {
	delegatedValidators, err := ioutil.ReadFile("./test_data/delegated_validators_response.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile("./test_data/validators_registry_validators_response.json")
	suite.NoError(err)

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/delegators/%s/validators", HubContract):       string(delegatedValidators),
		fmt.Sprintf("/wasm/contracts/%s/store", ValidatorsRegistryContract): string(whitelistedValidators),
	})
	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = "2"

	logger := NewTestLogger()
	valRepository := NewValidatorsRepository(cfg, logger)
	m := NewFailedRedelegationsMonitor(cfg, logger, valRepository)
	err = m.Handler(context.Background())
	suite.NoError(err)

	metricVectors := m.GetMetricVectors()

	expectedFailedValidatorsRedelegations := 1.0
	actualFailedValidatorsRedelegations := metricVectors[FailedRedelegations].Get(TestFailedRedelegationValidatorAddress)

	suite.Equal(expectedFailedValidatorsRedelegations, actualFailedValidatorsRedelegations)
}

func (suite *FailedRedelegationsMonitorTestSuite) TestRedelegationSucceedRequest() {
	delegatedValidators, err := ioutil.ReadFile("./test_data/delegated_validators_response_one_validator.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile("./test_data/validators_registry_validators_response.json")
	suite.NoError(err)

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/delegators/%s/validators", HubContract):       string(delegatedValidators),
		fmt.Sprintf("/wasm/contracts/%s/store", ValidatorsRegistryContract): string(whitelistedValidators),
	})
	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = "2"

	logger := NewTestLogger()
	valRepository := NewValidatorsRepository(cfg, logger)
	m := NewFailedRedelegationsMonitor(cfg, logger, valRepository)
	err = m.Handler(context.Background())
	suite.NoError(err)

	metricVectors := m.GetMetricVectors()

	failedValidatorsRedelegations := metricVectors[FailedRedelegations].Get(TestFailedRedelegationValidatorAddress)

	suite.Equal(failedValidatorsRedelegations, 0.0)
}
