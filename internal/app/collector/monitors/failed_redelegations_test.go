package monitors

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories"
	"github.com/lidofinance/terra-monitors/internal/app/collector/types"
	"github.com/lidofinance/terra-monitors/internal/pkg/stubs"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"

	"github.com/lidofinance/terra-repositories/delegations"

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
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	validatorInfoData, err := ioutil.ReadFile(dir + "test_data/columbus-5/slashing_validator_info_not_jailed.json")
	suite.NoError(err)

	delegatedValidators, err := ioutil.ReadFile(dir + "test_data/delegations_response.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile(dir + "test_data/validators_registry_validators_response.json")
	suite.NoError(err)

	testServer := stubs.NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", TestFailedRedelegationValidatorAddress):          string(validatorInfoData),
		fmt.Sprintf("/staking/validators/%s", TestDelegationValidatorAddressWithNonZeroShares): strings.Replace(string(validatorInfoData), types.TestMoniker, TestMoniker1, -1),
		fmt.Sprintf("/cosmos/staking/v1beta1/delegations/%s", types.HubContract):               string(delegatedValidators),
		fmt.Sprintf("/wasm/contracts/%s/store", types.ValidatorsRegistryContract):              string(whitelistedValidators),
	})
	cfg := stubs.NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = "2"
	cfg.NetworkGeneration = "columbus-5"
	logger := stubs.NewTestLogger()
	apiClient := utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger)

	valRepository, err := repositories.NewValidatorsRepository(stubs.BuildValidatorsRepositoryConfig(cfg), apiClient)
	suite.NoError(err)
	delRepository := delegations.New(apiClient)
	m := NewFailedRedelegationsMonitor(cfg, logger, valRepository, delRepository)
	err = m.Handler(context.Background())
	suite.NoError(err)

	metricVectors := m.GetMetricVectors()

	failedRedelegationValidatorLabel := fmt.Sprintf("%s (%s)", TestFailedRedelegationValidatorAddress, types.TestMoniker)
	delegationValidatorAddressWithNonZeroSharesLabel := fmt.Sprintf("%s (%s)", TestDelegationValidatorAddressWithNonZeroShares, TestMoniker1)

	expectedFailedValidatorsRedelegations := 1.0
	actualFailedValidatorsRedelegations := metricVectors[FailedRedelegations].Get(failedRedelegationValidatorLabel)
	suite.Equal(expectedFailedValidatorsRedelegations, actualFailedValidatorsRedelegations)

	actualFailedValidatorsRedelegations = metricVectors[FailedRedelegations].Get(delegationValidatorAddressWithNonZeroSharesLabel)
	suite.Equal(actualFailedValidatorsRedelegations, 0.0)
}

func (suite *FailedRedelegationsMonitorTestSuite) TestRedelegationSucceedRequest() {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	validatorInfoData, err := ioutil.ReadFile(dir + "test_data/columbus-5/slashing_validator_info_not_jailed.json")
	suite.NoError(err)

	delegatedValidators, err := ioutil.ReadFile(dir + "test_data/delegations_response_one_delegation.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile(dir + "test_data/validators_registry_validators_response.json")
	suite.NoError(err)

	testServer := stubs.NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", TestValidatorAddress):               string(validatorInfoData),
		fmt.Sprintf("/cosmos/staking/v1beta1/delegations/%s", types.HubContract):  string(delegatedValidators),
		fmt.Sprintf("/wasm/contracts/%s/store", types.ValidatorsRegistryContract): string(whitelistedValidators),
	})
	cfg := stubs.NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = "2"
	cfg.NetworkGeneration = "columbus-5"
	logger := stubs.NewTestLogger()
	apiClient := utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger)

	valRepository, err := repositories.NewValidatorsRepository(stubs.BuildValidatorsRepositoryConfig(cfg), apiClient)
	suite.NoError(err)
	delRepository := delegations.New(apiClient)
	m := NewFailedRedelegationsMonitor(cfg, logger, valRepository, delRepository)
	err = m.Handler(context.Background())
	suite.NoError(err)

	metricVectors := m.GetMetricVectors()

	label := fmt.Sprintf("%s (%s)", TestValidatorAddress, types.TestMoniker)
	failedValidatorsRedelegations := metricVectors[FailedRedelegations].Get(label)

	suite.Equal(0.0, failedValidatorsRedelegations)
}
