package monitors

import (
	"context"
	"fmt"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/sirupsen/logrus"
	"hash/crc32"
	"strings"
)

const (
	WhitelistedValidatorsCRC32 = "whitelisted_validators_crc32"
	WhitelistedValidatorsNum   = "whitelisted_validators_num"
)

type WhitelistedValidatorsMonitor struct {
	metrics   map[Metric]float64
	apiClient *client.TerraLiteForTerra
	logger    *logrus.Logger
	validatorsRepository ValidatorsRepository
}

func NewWhitelistedValidatorsMonitor(cfg config.CollectorConfig, repository ValidatorsRepository) WhitelistedValidatorsMonitor {
	m := WhitelistedValidatorsMonitor{
		metrics:   make(map[Metric]float64),
		apiClient: cfg.GetTerraClient(),
		logger:    cfg.Logger,
		validatorsRepository: repository,
	}

	return m
}

func (m WhitelistedValidatorsMonitor) Name() string {
	return "WhitelistedValidatorsMonitor"
}

func (m *WhitelistedValidatorsMonitor) InitMetrics() {
	m.metrics[WhitelistedValidatorsCRC32] = 0
	m.metrics[WhitelistedValidatorsNum] = 0
}

func (m WhitelistedValidatorsMonitor) GetMetrics() map[Metric]float64 {
	return m.metrics
}

func (m *WhitelistedValidatorsMonitor) Handler(ctx context.Context) error {

	validators,err := m.validatorsRepository.GetValidatorsAddresses(ctx)
	if err != nil {
		return fmt.Errorf("failed to get whiltelisted validators for %s: %w",m.Name(), err)
	}


	m.metrics[WhitelistedValidatorsCRC32] = float64(crc32.ChecksumIEEE([]byte(strings.Join(validators,""))))
	m.metrics[WhitelistedValidatorsNum] = float64(len(validators))
	m.logger.Infoln("updated ", m.Name())
	return nil
}
