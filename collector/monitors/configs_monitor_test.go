package monitors

import (
	"context"
	"encoding/json"

	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/stretchr/testify/suite"
)

type DetectorChangesTestSuite struct {
	suite.Suite
}

func (suite *DetectorChangesTestSuite) SetupTest() {

}

func (suite *DetectorChangesTestSuite) TestHubParameters() {
	hubParametersFakeResponse := struct {
		Height string
		Result interface{}
	}{Height: "100",
		Result: types.HubParameters{
			EpochPeriod:         10,
			UnderlyingCoinDenom: "uluna",
			UnbondingPeriod:     20,
			PegRecoveryFee:      "0.5",
			ErThreshold:         "1",
			RewardDenom:         "uusd",
		},
	}
	responseData, err := json.Marshal(hubParametersFakeResponse)
	suite.NoError(err)
	ts := NewServerWithResponse(string(responseData))
	cfg := NewTestCollectorConfig(ts.URL)

	logger := NewTestLogger()
	hubParametersMonitor := NewHubParametersMonitor(cfg, logger)

	err = hubParametersMonitor.Handler(context.Background())
	suite.NoError(err)
	suite.Equal(hubParametersFakeResponse.Result, *hubParametersMonitor.State)
	crc32first := hubParametersMonitor.metrics[HubParametersCRC32].Get()

	//   changing response data
	newHubParametersFakeResponse := struct {
		Height string
		Result interface{}
	}{Height: "100",
		Result: types.HubParameters{
			EpochPeriod:         100,
			UnderlyingCoinDenom: "uusd",
			UnbondingPeriod:     200,
			PegRecoveryFee:      "1.5",
			ErThreshold:         "2",
			RewardDenom:         "uluna",
		},
	}
	responseData, err = json.Marshal(newHubParametersFakeResponse)
	suite.NoError(err)

	ts = NewServerWithResponse(string(responseData))
	cfg = NewTestCollectorConfig(ts.URL)
	hubParametersMonitor = NewHubParametersMonitor(cfg, logger)

	err = hubParametersMonitor.Handler(context.Background())
	suite.NoError(err)
	suite.Equal(newHubParametersFakeResponse.Result, *hubParametersMonitor.State)

	crc32second := hubParametersMonitor.metrics[HubParametersCRC32]

	//detects crc32 is changed due to changed parameters data
	suite.NotEqual(crc32first, crc32second)
}

func (suite *DetectorChangesTestSuite) TestConfigsMonitorV2() {
	ts := NewServerWithRandomJson()
	cfg := NewTestCollectorConfig(ts.URL)
	cfg.BassetContractsVersion = "2"
	logger := NewTestLogger()
	m1 := NewConfigsCRC32Monitor(cfg, logger)
	savedMetrics := make(map[MetricName]MetricValue)

	err := m1.Handler(context.Background())
	suite.NoError(err)
	for metric, value := range m1.metrics {
		savedMetrics[metric] = &SimpleMetricValue{value: value.Get()}
	}
	err = m1.Handler(context.Background())
	suite.NoError(err)

	// since we are getting random data each http request
	// we should get m1.metrics and savedMetrics from first request differ each other
	suite.Equal(5, len(m1.metrics))
	suite.Equal(5, len(savedMetrics))
	for metric := range m1.metrics {
		_, found := savedMetrics[metric]
		suite.True(found)
		suite.NotEqual(m1.metrics[metric], savedMetrics[metric])
	}
}
