package monitors

import (
	"context"
	"fmt"
	"github.com/lidofinance/terra-monitors/collector/config"
	"io/ioutil"

	"github.com/stretchr/testify/suite"
)

type UpdateGlobalIndexMonitorTestSuite struct {
	suite.Suite
}

func (suite *UpdateGlobalIndexMonitorTestSuite) SetupTest() {
}

func (suite *UpdateGlobalIndexMonitorTestSuite) TestSuccessfulRequest() {
	expectedFailedTx := &ReadOnceMetric{value: 0.0}
	expectedSuccessTxs := &ReadOnceMetric{value: 10.0}
	// 1878948+1879023+1971021+1968141+1969755+1968301+1865889+1966868+1332487+1966896
	expectedGasUsed := &ReadOnceMetric{value: 18767329.0}
	// 2609420+2608384+2737567+2734520+2734575+2733759+2590159+2731335+2731705+2732807
	expectedGasWanted := &ReadOnceMetric{value: 26944231.0}
	// 391413+391258+410636+410178+410187+410064+388524+409701+409756+409922
	expectedUUSDUsed := &ReadOnceMetric{value: 4041639.0}

	data, err := ioutil.ReadFile("./test_data/columbus-5/update_global_index_success_response.json")
	suite.NoError(err)
	testServer := NewServerWithResponse(string(data))
	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.NetworkGeneration = config.NetworkGenerationColumbus5

	logger := NewTestLogger()
	m := NewUpdateGlobalIndexMonitor(cfg, logger)

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
	expectedFailedTx := &ReadOnceMetric{value: 1.0}
	expectedSuccessTxs := &ReadOnceMetric{value: 0.0}
	expectedGasUsed := &ReadOnceMetric{value: 1854478.0}
	expectedGasWanted := &ReadOnceMetric{value: 1842713.0}
	expectedUUSDUsed := &ReadOnceMetric{value: 276407.0}
	expectedErrorMessagePattern := "failed tx detected: out of gas: out of gas in location"

	data, err := ioutil.ReadFile("./test_data/columbus-5/update_global_index_error.json")
	suite.NoError(err)
	testServer := NewServerWithResponse(string(data))
	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.NetworkGeneration = config.NetworkGenerationColumbus5

	logger := NewTestLogger()
	m := NewUpdateGlobalIndexMonitor(cfg, logger)

	err = m.Handler(context.Background())
	suite.NoError(err)

	metrics := m.GetMetrics()

	suite.Equal(expectedFailedTx, metrics[UpdateGlobalIndexFailedTxSinceLastCheck])
	suite.Equal(expectedSuccessTxs, metrics[UpdateGlobalIndexSuccessfulTxSinceLastCheck])
	suite.Equal(expectedGasUsed, metrics[UpdateGlobalIndexGasUsed])
	suite.Equal(expectedGasWanted, metrics[UpdateGlobalIndexGasWanted])
	suite.Equal(expectedUUSDUsed, metrics[UpdateGlobalIndexUUSDFee])
	actualMessages := fmt.Sprintln(logger.Out)
	suite.Contains(actualMessages, expectedErrorMessagePattern)
}

func (suite *UpdateGlobalIndexMonitorTestSuite) testThresholdTxRequest(networkGeneration string) {
	expectedFailedTx := &ReadOnceMetric{value: 0.0}
	// tx counter is limited by threshold, 10 iteration, 10 tx each
	expectedSuccessTxs := &ReadOnceMetric{value: 100.0}
	expectedGasUsedPerTX := &ReadOnceMetric{value: 1000.0}
	expectedGasWantedPerTX := &ReadOnceMetric{value: 10000.0}
	expectedUUSDUsedPerTX := &ReadOnceMetric{value: 100000.0}
	expectedErrorMessagePattern := "update global index processing stopped due to requests threshold"

	testServer := NewServerForUpdateGlobalIndex(networkGeneration)
	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.NetworkGeneration = networkGeneration

	logger := NewTestLogger()
	m := NewUpdateGlobalIndexMonitor(cfg, logger)
	// by setting lastMaxCheckedID to some value, we are pretending its not a first run
	m.lastMaxCheckedID = 1

	err := m.Handler(context.Background())
	suite.NoError(err)

	metrics := m.GetMetrics()

	expectedSuccessTxsValue := expectedSuccessTxs.Get()
	suite.Equal(expectedFailedTx, metrics[UpdateGlobalIndexFailedTxSinceLastCheck])
	suite.Equal(expectedSuccessTxsValue, metrics[UpdateGlobalIndexSuccessfulTxSinceLastCheck].Get())
	suite.Equal(expectedSuccessTxsValue*expectedGasUsedPerTX.Get(), metrics[UpdateGlobalIndexGasUsed].Get())
	suite.Equal(expectedSuccessTxsValue*expectedGasWantedPerTX.Get(), metrics[UpdateGlobalIndexGasWanted].Get())
	suite.Equal(expectedSuccessTxsValue*expectedUUSDUsedPerTX.Get(), metrics[UpdateGlobalIndexUUSDFee].Get())
	actualMessages := fmt.Sprintln(logger.Out)
	suite.Contains(actualMessages, expectedErrorMessagePattern)
	suite.Equal(int64(200), m.lastMaxCheckedID)
}

func (suite *UpdateGlobalIndexMonitorTestSuite) TestThresholdTxRequest() {
	//suite.testThresholdTxRequest(config.NetworkGenerationColumbus5)
	suite.testThresholdTxRequest(config.NetworkGenerationColumbus4)
}

func (suite *UpdateGlobalIndexMonitorTestSuite) TestAlreadyCheckedTxRequest() {
	expectedFailedTx := &ReadOnceMetric{value: 0.0}
	expectedSuccessTxs := &ReadOnceMetric{value: 19.0}
	expectedGasUsedPerTX := &ReadOnceMetric{value: 1000.0}
	expectedGasWantedPerTX := &ReadOnceMetric{value: 10000.0}
	expectedUUSDUsedPerTX := &ReadOnceMetric{value: 100000.0}
	expectedErrorMessagePattern := "stopping processing, last checked transaction is found"

	testServer := NewServerForUpdateGlobalIndex(config.NetworkGenerationColumbus5)
	cfg := NewTestCollectorConfig(testServer.URL)
	cfg.NetworkGeneration = config.NetworkGenerationColumbus5

	logger := NewTestLogger()
	m := NewUpdateGlobalIndexMonitor(cfg, logger)
	// by setting lastMaxCheckedID to some value, we are pretending its not a first run
	m.lastMaxCheckedID = 181

	err := m.Handler(context.Background())
	suite.NoError(err)

	metrics := m.GetMetrics()

	expectedSuccessTxsValue := expectedSuccessTxs.Get()
	suite.Equal(expectedFailedTx, metrics[UpdateGlobalIndexFailedTxSinceLastCheck])
	suite.Equal(expectedSuccessTxsValue, metrics[UpdateGlobalIndexSuccessfulTxSinceLastCheck].Get())
	suite.Equal(expectedSuccessTxsValue*expectedGasUsedPerTX.Get(), metrics[UpdateGlobalIndexGasUsed].Get())
	suite.Equal(expectedSuccessTxsValue*expectedGasWantedPerTX.Get(), metrics[UpdateGlobalIndexGasWanted].Get())
	suite.Equal(expectedSuccessTxsValue*expectedUUSDUsedPerTX.Get(), metrics[UpdateGlobalIndexUUSDFee].Get())
	actualMessages := fmt.Sprintln(logger.Out)
	suite.Contains(actualMessages, expectedErrorMessagePattern)
	suite.Equal(int64(200), m.lastMaxCheckedID)
}
