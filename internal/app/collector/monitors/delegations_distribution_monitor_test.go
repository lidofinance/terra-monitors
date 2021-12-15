package monitors

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories"
	"github.com/lidofinance/terra-monitors/internal/app/collector/types"
	"github.com/lidofinance/terra-monitors/internal/pkg/stubs"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"

	"github.com/lidofinance/terra-repositories/delegations"

	"github.com/stretchr/testify/suite"
)

const (
	TestDelegationsDistributionOutlierLabel = "terravalcons1ezj3lps8nqwytt42at2sgt7seq9hk708g0sp09 (Test validator)"
)

type DelegationsDistributionTestSuite struct {
	suite.Suite
}

func (suite *DelegationsDistributionTestSuite) SetupTest() {

}

func (suite *DelegationsDistributionTestSuite) TestDelegationsDistributionNoPanic() {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	validatorInfoData, err := ioutil.ReadFile(dir + "/test_data/columbus-5/slashing_validator_info_not_jailed.json")
	suite.NoError(err)

	delegatedValidators, err := ioutil.ReadFile(dir + "/test_data/delegations_response_distribution_ok.json")
	suite.NoError(err)

	testServer := stubs.NewServerWithRoutedResponse(map[string]string{
		"/staking/validators/{address:[a-z0-9]+}":                                string(validatorInfoData),
		fmt.Sprintf("/cosmos/staking/v1beta1/delegations/%s", types.HubContract): string(delegatedValidators),
	})

	cfg := stubs.NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = "2"
	cfg.NetworkGeneration = "columbus-5"
	logger := stubs.NewTestLogger()
	apiClient := utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger)

	valRepository, err := repositories.NewValidatorsRepository(stubs.BuildValidatorsRepositoryConfig(cfg), apiClient)
	suite.NoError(err)
	delRepository := delegations.New(apiClient)
	m := NewDelegationsDistributionMonitor(cfg, logger, valRepository, delRepository)
	err = m.Handler(context.Background())
	suite.NoError(err)

	for _, value := range m.metricVectors[DelegationsDistributionImbalance].values {
		suite.Equal(float64(0), value)
	}
}

func (suite *DelegationsDistributionTestSuite) TestDelegationsDistributionPanic() {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	validatorInfoData, err := ioutil.ReadFile(dir + "test_data/columbus-5/slashing_validator_info_not_jailed.json")
	suite.NoError(err)

	delegatedValidators, err := ioutil.ReadFile(dir + "test_data/delegations_response_distribution_not_ok.json")
	suite.NoError(err)

	testServer := stubs.NewServerWithRoutedResponse(map[string]string{
		"/staking/validators/{address:[a-z0-9]+}":                                string(validatorInfoData),
		fmt.Sprintf("/cosmos/staking/v1beta1/delegations/%s", types.HubContract): string(delegatedValidators),
	})

	cfg := stubs.NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = "2"
	cfg.NetworkGeneration = "columbus-5"
	logger := stubs.NewTestLogger()
	apiClient := utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger)

	valRepository, err := repositories.NewValidatorsRepository(stubs.BuildValidatorsRepositoryConfig(cfg), apiClient)
	suite.NoError(err)
	delRepository := delegations.New(apiClient)
	m := NewDelegationsDistributionMonitor(cfg, logger, valRepository, delRepository)
	err = m.Handler(context.Background())
	suite.NoError(err)

	for idx, value := range m.metricVectors[DelegationsDistributionImbalance].values {
		if idx == TestDelegationsDistributionOutlierLabel {
			suite.Equal(float64(1), value)
		} else {
			suite.Equal(float64(0), value)
		}
	}
}
