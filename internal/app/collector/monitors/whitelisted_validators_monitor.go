package monitors

import (
	"context"
	"fmt"
	"hash/crc32"
	"sort"
	"strings"

	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories/validators"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/client"
	terraClient "github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/sirupsen/logrus"
)

const (
	WhitelistedValidatorsCRC32 MetricName = "whitelisted_validators_crc32"
	WhitelistedValidatorsNum   MetricName = "whitelisted_validators_num"
)

type WhitelistedValidatorsMonitor struct {
	metrics              map[MetricName]MetricValue
	apiClient            *terraClient.TerraLiteForTerra
	logger               *logrus.Logger
	validatorsRepository validators.ValidatorsRepository
}

func NewWhitelistedValidatorsMonitor(
	cfg config.CollectorConfig,
	logger *logrus.Logger,
	repository validators.ValidatorsRepository,
) WhitelistedValidatorsMonitor {
	m := WhitelistedValidatorsMonitor{
		metrics:              make(map[MetricName]MetricValue),
		apiClient:            client.New(cfg.LCD, logger),
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

func (m *WhitelistedValidatorsMonitor) InitMetrics() {
	for _, metric := range m.providedMetrics() {
		if m.metrics[metric] == nil {
			m.metrics[metric] = &SimpleMetricValue{}
		}
		m.metrics[metric].Set(0)
	}
}

func (m WhitelistedValidatorsMonitor) GetMetrics() map[MetricName]MetricValue {
	return m.metrics
}

func (m WhitelistedValidatorsMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	return nil
}

func (m *WhitelistedValidatorsMonitor) Handler(ctx context.Context) error {

	validators, err := m.validatorsRepository.GetValidatorsAddresses(ctx)
	if err != nil {
		return fmt.Errorf("failed to get whiltelisted validators for %s: %w", m.Name(), err)
	}

	sort.Strings(validators)
	m.metrics[WhitelistedValidatorsCRC32].Set(float64(crc32.ChecksumIEEE([]byte(strings.Join(validators, "")))))
	m.metrics[WhitelistedValidatorsNum].Set(float64(len(validators)))
	m.logger.Infoln("updated ", m.Name())
	return nil
}
