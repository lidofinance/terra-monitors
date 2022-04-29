package monitors

import (
	"context"
	"fmt"
	"hash/crc32"
	"sort"
	"strings"

	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"

	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client"

	"github.com/sirupsen/logrus"
)

const (
	WhitelistedValidatorsCRC32 MetricName = "whitelisted_validators_crc32"
	WhitelistedValidatorsNum   MetricName = "whitelisted_validators_num"
	WhitelistedValidators      MetricName = "whitelisted_validators"
)

type WhitelistedValidatorsMonitor struct {
	metrics              map[MetricName]MetricValue
	metricVectors        map[MetricName]*MetricVector
	apiClient            *client.TerraRESTApis
	logger               *logrus.Logger
	validatorsRepository repositories.ValidatorsRepository
}

func NewWhitelistedValidatorsMonitor(
	cfg config.CollectorConfig,
	logger *logrus.Logger,
	repository repositories.ValidatorsRepository,
) WhitelistedValidatorsMonitor {
	m := WhitelistedValidatorsMonitor{
		metrics:              make(map[MetricName]MetricValue),
		metricVectors:        make(map[MetricName]*MetricVector),
		apiClient:            utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger),
		logger:               logger,
		validatorsRepository: repository,
	}
	m.InitMetrics()

	return m
}

func (m WhitelistedValidatorsMonitor) Name() string {
	return "WhitelistedValidatorsMonitor"
}

func (m *WhitelistedValidatorsMonitor) providedMetrics() []MetricName {
	return []MetricName{
		WhitelistedValidatorsCRC32,
		WhitelistedValidatorsNum,
	}
}

func (m *WhitelistedValidatorsMonitor) providedMetricVectors() []MetricName {
	return []MetricName{
		WhitelistedValidators,
	}
}

func (m *WhitelistedValidatorsMonitor) InitMetrics() {
	initMetrics(m.providedMetrics(), m.providedMetricVectors(), m.metrics, m.metricVectors)
}

func (m WhitelistedValidatorsMonitor) GetMetrics() map[MetricName]MetricValue {
	return m.metrics
}

func (m WhitelistedValidatorsMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	return m.metricVectors
}

func (m *WhitelistedValidatorsMonitor) Handler(ctx context.Context) error {

	validators, err := m.validatorsRepository.GetValidatorsAddresses(ctx)
	if err != nil {
		return fmt.Errorf("failed to get whiltelisted validators for %s: %w", m.Name(), err)
	}

	sort.Strings(validators)
	m.metrics[WhitelistedValidatorsCRC32].Set(float64(crc32.ChecksumIEEE([]byte(strings.Join(validators, "")))))
	m.metrics[WhitelistedValidatorsNum].Set(float64(len(validators)))

	for _, validator := range validators {
		info, err := m.validatorsRepository.GetValidatorInfo(ctx, validator)
		if err != nil {
			return fmt.Errorf("failed to get whiltelisted validator %s info: %w", validator, err)
		}

		m.metricVectors[WhitelistedValidators].Set(info.Moniker, 1)
	}

	m.logger.Infoln("updated", m.Name())
	return nil
}
