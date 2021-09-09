package monitors

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/stretchr/testify/suite"
)

type OracleVotesMonitorTestSuite struct {
	suite.Suite
}

func (suite *OracleVotesMonitorTestSuite) SetupTest() {

}

func (suite *OracleVotesMonitorTestSuite) TestSuccessfulRequest() {
	validatorInfoData, err := ioutil.ReadFile("./test_data/slashing_validator_info_not_jailed.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile("./test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	oracleParams, err := ioutil.ReadFile("./test_data/oracle_parameters.json")
	suite.NoError(err)

	oracleMissedVotePeriods, err := ioutil.ReadFile("./test_data/oracle_missed_vote_periods.json")
	suite.NoError(err)

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", testValAddress): string(validatorInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", HubContract):  string(whitelistedValidators),
		fmt.Sprintf("/oracle/parameters"):                     string(oracleParams),
		fmt.Sprintf("/oracle/voters/%s/miss", testValAddress): string(oracleMissedVotePeriods),
	})
	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = "1"

	logger := NewTestLogger()
	valRepository := NewValidatorsRepository(cfg, logger)
	m := NewOracleVotesMonitor(cfg, logger, valRepository)
	err = m.Handler(context.Background())
	suite.NoError(err)

	metricVectors := m.GetMetricVectors()

	expectedValidatorsCommission := 0.1
	actualValidatorsCommission := metricVectors[OracleMissedVoteRate].Get(TestMoniker)

	suite.Equal(expectedValidatorsCommission, actualValidatorsCommission)
}

func (suite *OracleVotesMonitorTestSuite) TestFailedValidatorsFeeRequest() {
	validatorInfoData, err := ioutil.ReadFile("./test_data/slashing_validator_info_not_jailed.json")
	suite.NoError(err)

	whitelistedValidators, err := ioutil.ReadFile("./test_data/whitelisted_validators_response.json")
	suite.NoError(err)

	oracleParams, err := ioutil.ReadFile("./test_data/slashing_error.json")
	suite.NoError(err)

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", testValAddress): string(validatorInfoData),
		fmt.Sprintf("/wasm/contracts/%s/store", HubContract):  string(whitelistedValidators),
		fmt.Sprintf("/oracle/parameters"):                     string(oracleParams),
	})
	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.BassetContractsVersion = "1"

	logger := NewTestLogger()
	valRepository := NewValidatorsRepository(cfg, logger)
	m := NewOracleVotesMonitor(cfg, logger, valRepository)
	err = m.Handler(context.Background())
	suite.Error(err)
}
