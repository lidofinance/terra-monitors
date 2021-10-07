package monitors

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/lidofinance/terra-monitors/collector/monitors/delegations"

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
	validatorInfoData, err := ioutil.ReadFile(fmt.Sprintf("./test_data/columbus-5/slashing_validator_info_not_jailed.json"))
	suite.NoError(err)

	delegatedValidators, err := ioutil.ReadFile("./test_data/delegations_response_distribution_ok.json")
	suite.NoError(err)

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/{address:[a-z0-9]+}"):             string(validatorInfoData),
		fmt.Sprintf("/cosmos/staking/v1beta1/delegations/%s", HubContract): string(delegatedValidators),
	})

	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = "2"
	cfg.NetworkGeneration = "columbus-5"

	logger := NewTestLogger()
	valRepository := NewValidatorsRepository(cfg, logger)
	delRepository := delegations.New(cfg, logger)
	m := NewDelegationsDistributionMonitor(cfg, logger, valRepository, delRepository)
	err = m.Handler(context.Background())
	suite.NoError(err)

	for _, value := range m.metricVectors[DelegationsDistributionImbalance].values {
		suite.Equal(float64(0), value)
	}
}

func (suite *DelegationsDistributionTestSuite) TestDelegationsDistributionPanic() {
	validatorInfoData, err := ioutil.ReadFile(fmt.Sprintf("./test_data/columbus-5/slashing_validator_info_not_jailed.json"))
	suite.NoError(err)

	delegatedValidators, err := ioutil.ReadFile("./test_data/delegations_response_distribution_not_ok.json")
	suite.NoError(err)

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/{address:[a-z0-9]+}"):             string(validatorInfoData),
		fmt.Sprintf("/cosmos/staking/v1beta1/delegations/%s", HubContract): string(delegatedValidators),
	})

	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = "2"
	cfg.NetworkGeneration = "columbus-5"

	logger := NewTestLogger()
	valRepository := NewValidatorsRepository(cfg, logger)
	delRepository := delegations.New(cfg, logger)
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
