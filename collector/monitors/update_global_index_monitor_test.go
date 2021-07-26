package monitors

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/stretchr/testify/suite"
)

type UpdateGlobalIndexMonitorTestSuite struct {
	suite.Suite
}

func (suite *UpdateGlobalIndexMonitorTestSuite) SetupTest() {

}

func (suite *UpdateGlobalIndexMonitorTestSuite) TestSuccessfulRequest() {
	expectedFailedTx := 0.0
	expectedSuccessTxs := 10.0
	// 1878948+1879023+1971021+1968141+1969755+1968301+1865889+1966868+1332487+1966896
	expectedGasUsed := 18767329.0
	// 2609420+2608384+2737567+2734520+2734575+2733759+2590159+2731335+2731705+2732807
	expectedGasWanted := 26944231.0
	// 391413+391258+410636+410178+410187+410064+388524+409701+409756+409922
	expectedUUSDUsed := 4041639.0

	data, err := ioutil.ReadFile("./test_data/update_global_index_success_response.json")
	suite.NoError(err)
	testServer := NewServerWithResponse(string(data))
	cfg := NewTestCollectorConfig(testServer.URL)

	m := NewUpdateGlobalIndexMonitor(cfg)
	err = m.Handler(context.Background())
	suite.NoError(err)

	metrics := m.GetMetrics()

	suite.Equal(expectedFailedTx, metrics[UpdateGlobalIndexFailedTxSinceLastCheck])
	suite.Equal(expectedSuccessTxs, metrics[UpdateGlobalIndexSuccessfulTxSinceLastCheck])
	suite.Equal(expectedGasUsed, metrics[UpdateGlobalIndexGasUsed])
	suite.Equal(expectedGasWanted, metrics[UpdateGlobalIndexGasWanted])
	suite.Equal(expectedUUSDUsed, metrics[UpdateGlobalIndexUUSDFee])
}

func (suite *UpdateGlobalIndexMonitorTestSuite) TestFailedTxRequest() {
	expectedFailedTx := 1.0
	expectedSuccessTxs := 0.0
	expectedGasUsed := 1854478.0
	expectedGasWanted := 1842713.0
	expectedUUSDUsed := 276407.0
	expectedErrorMessagePattern := "failed tx detected: out of gas: out of gas in location"

	data, err := ioutil.ReadFile("./test_data/update_global_index_error.json")
	suite.NoError(err)
	testServer := NewServerWithResponse(string(data))
	cfg := NewTestCollectorConfig(testServer.URL)

	m := NewUpdateGlobalIndexMonitor(cfg)
	err = m.Handler(context.Background())
	suite.NoError(err)

	metrics := m.GetMetrics()

	suite.Equal(expectedFailedTx, metrics[UpdateGlobalIndexFailedTxSinceLastCheck])
	suite.Equal(expectedSuccessTxs, metrics[UpdateGlobalIndexSuccessfulTxSinceLastCheck])
	suite.Equal(expectedGasUsed, metrics[UpdateGlobalIndexGasUsed])
	suite.Equal(expectedGasWanted, metrics[UpdateGlobalIndexGasWanted])
	suite.Equal(expectedUUSDUsed, metrics[UpdateGlobalIndexUUSDFee])
	actualMessages := fmt.Sprintln(cfg.Logger.Out)
	suite.Contains(actualMessages, expectedErrorMessagePattern)
}

func (suite *UpdateGlobalIndexMonitorTestSuite) TestThresholdTxRequest() {
	expectedFailedTx := 0.0
	// tx counter is limited by threshold, 10 iteration, 10 tx each
	expectedSuccessTxs := 100.0
	expectedGasUsedPerTX := 1000.0
	expectedGasWantedPerTX := 10000.0
	expectedUUSDUsedPerTX := 100000.0
	expectedErrorMessagePattern := "update global index processing stopped due to requests threshold"

	testServer := NewServerForUpdateGlobalIndex()
	cfg := NewTestCollectorConfig(testServer.URL)

	m := NewUpdateGlobalIndexMonitor(cfg)
	// by setting lastMaxCheckedID to some value, we are pretending its not a first run
	m.lastMaxCheckedID = 1
	err := m.Handler(context.Background())
	suite.NoError(err)

	metrics := m.GetMetrics()

	suite.Equal(expectedFailedTx, metrics[UpdateGlobalIndexFailedTxSinceLastCheck])
	suite.Equal(expectedSuccessTxs, metrics[UpdateGlobalIndexSuccessfulTxSinceLastCheck])
	suite.Equal(expectedSuccessTxs*expectedGasUsedPerTX, metrics[UpdateGlobalIndexGasUsed])
	suite.Equal(expectedSuccessTxs*expectedGasWantedPerTX, metrics[UpdateGlobalIndexGasWanted])
	suite.Equal(expectedSuccessTxs*expectedUUSDUsedPerTX, metrics[UpdateGlobalIndexUUSDFee])
	actualMessages := fmt.Sprintln(cfg.Logger.Out)
	suite.Contains(actualMessages, expectedErrorMessagePattern)
}

func (suite *UpdateGlobalIndexMonitorTestSuite) TestAlreadyCheckedTxRequest() {
	expectedFailedTx := 0.0
	expectedSuccessTxs := 19.0
	expectedGasUsedPerTX := 1000.0
	expectedGasWantedPerTX := 10000.0
	expectedUUSDUsedPerTX := 100000.0
	expectedErrorMessagePattern := "stopping processing, last checked transaction is found"

	testServer := NewServerForUpdateGlobalIndex()
	cfg := NewTestCollectorConfig(testServer.URL)

	m := NewUpdateGlobalIndexMonitor(cfg)
	// by setting lastMaxCheckedID to some value, we are pretending its not a first run
	m.lastMaxCheckedID = 181
	err := m.Handler(context.Background())
	suite.NoError(err)

	metrics := m.GetMetrics()

	suite.Equal(expectedFailedTx, metrics[UpdateGlobalIndexFailedTxSinceLastCheck])
	suite.Equal(expectedSuccessTxs, metrics[UpdateGlobalIndexSuccessfulTxSinceLastCheck])
	suite.Equal(expectedSuccessTxs*expectedGasUsedPerTX, metrics[UpdateGlobalIndexGasUsed])
	suite.Equal(expectedSuccessTxs*expectedGasWantedPerTX, metrics[UpdateGlobalIndexGasWanted])
	suite.Equal(expectedSuccessTxs*expectedUUSDUsedPerTX, metrics[UpdateGlobalIndexUUSDFee])
	actualMessages := fmt.Sprintln(cfg.Logger.Out)
	suite.Contains(actualMessages, expectedErrorMessagePattern)
}
