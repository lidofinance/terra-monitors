package monitors

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/stretchr/testify/suite"
)

const (
	testValAddress   = "terravalcons1ezj3lps8nqwytt42at2sgt7seq9hk708g0spyk"
	testValPublicKey = "terravalconspub1zcjduepqw2hyr7u7y70z5kdewn00xuq0wwcvnn0s7x5pjqcdpn80qsyctcpqcjhz4c"
)

type SlashingMonitorTestSuite struct {
	suite.Suite
}

func (suite *SlashingMonitorTestSuite) SetupTest() {

}

type MockV1ValidatorsRepository struct{}

func (r *MockV1ValidatorsRepository) GetValidatorsAddresses(ctx context.Context) ([]string, error) {
	return []string{
		testValAddress,
	}, nil
}

func (suite *SlashingMonitorTestSuite) TestSuccessfulRequestWithSlashing() {
	validatorInfoData, err := ioutil.ReadFile("./test_data/slashing_validator_info.json")
	suite.NoError(err)

	validatorSigningInfoData, err := ioutil.ReadFile(
		"./test_data/slashing_success_response_blocks_jailed_tombstoned.json")
	suite.NoError(err)

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", testValAddress):                 string(validatorInfoData),
		fmt.Sprintf("/slashing/validators/%s/signing_info", testValPublicKey): string(validatorSigningInfoData),
	})
	cfg := NewTestCollectorConfig(testServer.URL)

	m := NewSlashingMonitor(cfg, &MockV1ValidatorsRepository{})
	err = m.Handler(context.Background())
	suite.NoError(err)

	metrics := m.GetMetrics()
	metricVectors := m.GetMetricVectors()
	var (
		expectedNumTombstonedValidators float64 = 1
		expectedNumJailedValidators     float64 = 1
		expectedNumMissedBlocks         float64 = 5
	)
	var actualMissedBlocks float64
	for _,missedBlocks := range metricVectors[SlashingNumMissedBlocks] {
		actualMissedBlocks += missedBlocks
	}
	suite.Equal(expectedNumTombstonedValidators, metrics[SlashingNumTombstonedValidators])
	suite.Equal(expectedNumJailedValidators, metrics[SlashingNumJailedValidators])
	suite.Equal(expectedNumMissedBlocks, actualMissedBlocks)
}

func (suite *SlashingMonitorTestSuite) TestSuccessfulRequestNoSlashing() {
	validatorInfoData, err := ioutil.ReadFile("./test_data/slashing_validator_info.json")
	suite.NoError(err)

	validatorSigningInfoData, err := ioutil.ReadFile(
		"./test_data/slashing_success_response_no_slashing.json")
	suite.NoError(err)

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", testValAddress):                 string(validatorInfoData),
		fmt.Sprintf("/slashing/validators/%s/signing_info", testValPublicKey): string(validatorSigningInfoData),
	})
	cfg := NewTestCollectorConfig(testServer.URL)

	m := NewSlashingMonitor(cfg, &MockV1ValidatorsRepository{})
	err = m.Handler(context.Background())
	suite.NoError(err)

	metrics := m.GetMetrics()
	metricVectors := m.GetMetricVectors()

	var (
		expectedNumTombstonedValidators float64 = 0
		expectedNumJailedValidators     float64 = 0
		expectedNumMissedBlocks         float64 = 0
	)
	var actualMissedBlocks float64
	for _,missedBlocks := range metricVectors[SlashingNumMissedBlocks] {
		actualMissedBlocks += missedBlocks
	}
	suite.Equal(expectedNumTombstonedValidators, metrics[SlashingNumTombstonedValidators])
	suite.Equal(expectedNumJailedValidators, metrics[SlashingNumJailedValidators])
	suite.Equal(expectedNumMissedBlocks, actualMissedBlocks)
}

func (suite *UpdateGlobalIndexMonitorTestSuite) TestFailedSlashingRequest() {
	validatorInfoData, err := ioutil.ReadFile("./test_data/slashing_error.json")
	suite.NoError(err)

	validatorSigningInfoData, err := ioutil.ReadFile(
		"./test_data/slashing_success_response_blocks_jailed_tombstoned.json")
	suite.NoError(err)

	testServer := NewServerWithRoutedResponse(map[string]string{
		fmt.Sprintf("/staking/validators/%s", testValAddress):                 string(validatorInfoData),
		fmt.Sprintf("/slashing/validators/%s/signing_info", testValPublicKey): string(validatorSigningInfoData),
	})
	cfg := NewTestCollectorConfig(testServer.URL)

	m := NewSlashingMonitor(cfg, &MockV1ValidatorsRepository{})
	err = m.Handler(context.Background())
	suite.Error(err)
}
