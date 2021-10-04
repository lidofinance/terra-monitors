package monitors

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/stretchr/testify/suite"
)

const (
	TestFailedRedelegationValidatorAddress          = "terravaloper1qxqrtvg3smlfdfhvwcdzh0huh4f50kfs6gdt4x"
	TestDelegationValidatorAddressWithNonZeroShares = "terravaloper1qxqrtvg3smlfdfhvwcdzh0huh4f50kfs6gdg38"
	TestValidatorAddress                            = "terravalcons1ezj3lps8nqwytt42at2sgt7seq9hk708g0spyk"
	TestMoniker1                                    = "Test validator1"
)

type FailedRedelegationsMonitorTestSuite struct {
	suite.Suite
}

func (suite *FailedRedelegationsMonitorTestSuite) SetupTest() {

}

func (suite *FailedRedelegationsMonitorTestSuite) TestRedelegationFailedRequest() {
	validatorInfoData, err := ioutil.ReadFile(fmt.Sprintf("./test_data/columbus-5/slashing_validator_info_not_jailed.json"))
	suite.NoError(err)

	delegatedValidators, err := ioutil.ReadFile("./test_data/delegations_response.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile("./test_data/validators_registry_validators_response.json")
	suite.NoError(err)

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", TestFailedRedelegationValidatorAddress):          string(validatorInfoData),
		fmt.Sprintf("/staking/validators/%s", TestDelegationValidatorAddressWithNonZeroShares): strings.Replace(string(validatorInfoData), TestMoniker, TestMoniker1, -1),
		fmt.Sprintf("/cosmos/staking/v1beta1/delegations/%s", HubContract):                     string(delegatedValidators),
		fmt.Sprintf("/wasm/contracts/%s/store", ValidatorsRegistryContract):                    string(whitelistedValidators),
	})
	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = "2"
	cfg.NetworkGeneration = "columbus-5"

	logger := NewTestLogger()
	valRepository := NewValidatorsRepository(cfg, logger)
	m := NewFailedRedelegationsMonitor(cfg, logger, valRepository)
	err = m.Handler(context.Background())
	suite.NoError(err)

	metricVectors := m.GetMetricVectors()

	failedRedelegationValidatorLabel := fmt.Sprintf("%s (%s)", TestFailedRedelegationValidatorAddress, TestMoniker)
	delegationValidatorAddressWithNonZeroSharesLabel := fmt.Sprintf("%s (%s)", TestDelegationValidatorAddressWithNonZeroShares, TestMoniker1)

	expectedFailedValidatorsRedelegations := 1.0
	actualFailedValidatorsRedelegations := metricVectors[FailedRedelegations].Get(failedRedelegationValidatorLabel)
	suite.Equal(expectedFailedValidatorsRedelegations, actualFailedValidatorsRedelegations)

	actualFailedValidatorsRedelegations = metricVectors[FailedRedelegations].Get(delegationValidatorAddressWithNonZeroSharesLabel)
	suite.Equal(actualFailedValidatorsRedelegations, 0.0)
}

func (suite *FailedRedelegationsMonitorTestSuite) TestRedelegationSucceedRequest() {
	validatorInfoData, err := ioutil.ReadFile(fmt.Sprintf("./test_data/columbus-5/slashing_validator_info_not_jailed.json"))
	suite.NoError(err)

	delegatedValidators, err := ioutil.ReadFile("./test_data/delegations_response_one_delegation.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile("./test_data/validators_registry_validators_response.json")
	suite.NoError(err)

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", TestValidatorAddress):         string(validatorInfoData),
		fmt.Sprintf("/cosmos/staking/v1beta1/delegations/%s", HubContract):  string(delegatedValidators),
		fmt.Sprintf("/wasm/contracts/%s/store", ValidatorsRegistryContract): string(whitelistedValidators),
	})
	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = "2"
	cfg.NetworkGeneration = "columbus-5"

	logger := NewTestLogger()
	valRepository := NewValidatorsRepository(cfg, logger)
	m := NewFailedRedelegationsMonitor(cfg, logger, valRepository)
	err = m.Handler(context.Background())
	suite.NoError(err)

	metricVectors := m.GetMetricVectors()

	label := fmt.Sprintf("%s (%s)", TestValidatorAddress, TestMoniker)
	failedValidatorsRedelegations := metricVectors[FailedRedelegations].Get(label)

	suite.Equal(0.0, failedValidatorsRedelegations)
}
