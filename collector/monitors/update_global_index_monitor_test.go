package monitors

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"time"
)

type UpdateGlobalIndexMonitorTestSuite struct {
	suite.Suite
}

func (suite *UpdateGlobalIndexMonitorTestSuite) SetupTest() {

}

func (suite *UpdateGlobalIndexMonitorTestSuite) TestSuccessfulRequest() {
	expectedFailedTx := &ReadOnceMetric{0.0}
	expectedSuccessTxs := &ReadOnceMetric{10.0}
	// 1878948+1879023+1971021+1968141+1969755+1968301+1865889+1966868+1332487+1966896
	expectedGasUsed := &ReadOnceMetric{18767329.0}
	// 2609420+2608384+2737567+2734520+2734575+2733759+2590159+2731335+2731705+2732807
	expectedGasWanted := &ReadOnceMetric{26944231.0}
	// 391413+391258+410636+410178+410187+410064+388524+409701+409756+409922
	expectedUUSDUsed := &ReadOnceMetric{4041639.0}

	data, err := ioutil.ReadFile("./test_data/update_global_index_success_response.json")
	suite.NoError(err)
	testServer := NewServerWithResponse(string(data))
	cfg := NewTestCollectorConfig(testServer.URL)

	m := NewUpdateGlobalIndexMonitor(cfg)
	// m.Handler(ctx) does nothing, so we are forcing monitor to update data via flowManager channel
	m.flowManager <- struct{}{}
	//give a time to goroutine to make work
	time.Sleep(100 * time.Millisecond)

	metrics := m.GetMetrics()

	suite.Equal(expectedFailedTx, metrics[UpdateGlobalIndexFailedTxSinceLastCheck])
	suite.Equal(expectedSuccessTxs, metrics[UpdateGlobalIndexSuccessfulTxSinceLastCheck])
	suite.Equal(expectedGasUsed, metrics[UpdateGlobalIndexGasUsed])
	suite.Equal(expectedGasWanted, metrics[UpdateGlobalIndexGasWanted])
	suite.Equal(expectedUUSDUsed, metrics[UpdateGlobalIndexUUSDFee])
}

func (suite *UpdateGlobalIndexMonitorTestSuite) TestFailedTxRequest() {
	expectedFailedTx := &ReadOnceMetric{1.0}
	expectedSuccessTxs := &ReadOnceMetric{0.0}
	expectedGasUsed := &ReadOnceMetric{1854478.0}
	expectedGasWanted := &ReadOnceMetric{1842713.0}
	expectedUUSDUsed := &ReadOnceMetric{276407.0}
	expectedErrorMessagePattern := "failed tx detected: out of gas: out of gas in location"

	data, err := ioutil.ReadFile("./test_data/update_global_index_error.json")
	suite.NoError(err)
	testServer := NewServerWithResponse(string(data))
	cfg := NewTestCollectorConfig(testServer.URL)

	m := NewUpdateGlobalIndexMonitor(cfg)
	// m.Handler(ctx) does nothing, so we are forcing monitor to update data via flowManager channel
	m.flowManager <- struct{}{}
	//give a time to goroutine to make work
	time.Sleep(100 * time.Millisecond)

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
	expectedFailedTx := &ReadOnceMetric{0.0}
	// tx counter is limited by threshold, 10 iteration, 10 tx each
	expectedSuccessTxs := &ReadOnceMetric{100.0}
	expectedGasUsedPerTX := &ReadOnceMetric{1000.0}
	expectedGasWantedPerTX := &ReadOnceMetric{10000.0}
	expectedUUSDUsedPerTX := &ReadOnceMetric{100000.0}
	expectedErrorMessagePattern := "update global index processing stopped due to requests threshold"

	testServer := NewServerForUpdateGlobalIndex()
	cfg := NewTestCollectorConfig(testServer.URL)

	m := NewUpdateGlobalIndexMonitor(cfg)
	// by setting lastMaxCheckedID to some value, we are pretending its not a first run
	m.lastMaxCheckedID = 1

	// m.Handler(ctx) does nothing, so we are forcing monitor to update data via flowManager channel
	m.flowManager <- struct{}{}
	//give a time to goroutine to make work
	time.Sleep(100 * time.Millisecond)

	metrics := m.GetMetrics()

	expectedSuccessTxsValue := expectedSuccessTxs.Get()
	suite.Equal(expectedFailedTx, metrics[UpdateGlobalIndexFailedTxSinceLastCheck])
	suite.Equal(expectedSuccessTxsValue, metrics[UpdateGlobalIndexSuccessfulTxSinceLastCheck].Get())
	suite.Equal(expectedSuccessTxsValue*expectedGasUsedPerTX.Get(), metrics[UpdateGlobalIndexGasUsed].Get())
	suite.Equal(expectedSuccessTxsValue*expectedGasWantedPerTX.Get(), metrics[UpdateGlobalIndexGasWanted].Get())
	suite.Equal(expectedSuccessTxsValue*expectedUUSDUsedPerTX.Get(), metrics[UpdateGlobalIndexUUSDFee].Get())
	actualMessages := fmt.Sprintln(cfg.Logger.Out)
	suite.Contains(actualMessages, expectedErrorMessagePattern)
	suite.Equal(200, m.lastMaxCheckedID)
}

func (suite *UpdateGlobalIndexMonitorTestSuite) TestAlreadyCheckedTxRequest() {
	expectedFailedTx := &ReadOnceMetric{0.0}
	expectedSuccessTxs := &ReadOnceMetric{19.0}
	expectedGasUsedPerTX := &ReadOnceMetric{1000.0}
	expectedGasWantedPerTX := &ReadOnceMetric{10000.0}
	expectedUUSDUsedPerTX := &ReadOnceMetric{100000.0}
	expectedErrorMessagePattern := "stopping processing, last checked transaction is found"

	testServer := NewServerForUpdateGlobalIndex()
	cfg := NewTestCollectorConfig(testServer.URL)

	m := NewUpdateGlobalIndexMonitor(cfg)
	// by setting lastMaxCheckedID to some value, we are pretending its not a first run
	m.lastMaxCheckedID = 181
	// m.Handler(ctx) does nothing, so we are forcing monitor to update data via flowManager channel
	m.flowManager <- struct{}{}
	//give a time to goroutine to make work
	time.Sleep(100 * time.Millisecond)

	metrics := m.GetMetrics()

	expectedSuccessTxsValue := expectedSuccessTxs.Get()
	suite.Equal(expectedFailedTx, metrics[UpdateGlobalIndexFailedTxSinceLastCheck])
	suite.Equal(expectedSuccessTxsValue, metrics[UpdateGlobalIndexSuccessfulTxSinceLastCheck].Get())
	suite.Equal(expectedSuccessTxsValue*expectedGasUsedPerTX.Get(), metrics[UpdateGlobalIndexGasUsed].Get())
	suite.Equal(expectedSuccessTxsValue*expectedGasWantedPerTX.Get(), metrics[UpdateGlobalIndexGasWanted].Get())
	suite.Equal(expectedSuccessTxsValue*expectedUUSDUsedPerTX.Get(), metrics[UpdateGlobalIndexUUSDFee].Get())
	actualMessages := fmt.Sprintln(cfg.Logger.Out)
	suite.Contains(actualMessages, expectedErrorMessagePattern)
	suite.Equal(200, m.lastMaxCheckedID)
}
