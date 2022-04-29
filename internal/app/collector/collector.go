package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/lidofinance/terra-monitors/internal/app/collector/monitors"
	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"

	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client"
	"github.com/lidofinance/terra-repositories/delegations"
	"github.com/lidofinance/terra-repositories/signinfo"

	"github.com/sirupsen/logrus"
)

func New(cfg config.CollectorConfig, logger *logrus.Logger) (*Collector, error) {
	c := &Collector{
		Metrics:       make(map[monitors.MetricName]monitors.Monitor),
		MetricVectors: make(map[monitors.MetricName]monitors.Monitor),
		logger:        logger,
		apiClient:     utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger),
	}
	ctx := context.Background()

	hubStateMonitor := monitors.NewHubStateMonitor(cfg, logger)
	c.RegisterMonitor(ctx, cfg, hubStateMonitor)

	rewardStateMonitor := monitors.NewRewardStateMonitor(cfg, logger)
	c.RegisterMonitor(ctx, cfg, &rewardStateMonitor)

	blunaTokenInfoMonitor := monitors.NewBlunaTokenInfoMonitor(cfg, logger)
	c.RegisterMonitor(ctx, cfg, blunaTokenInfoMonitor)

	valRepoCfg := repositories.ValidatorsRepositoryConfig{
		BAssetContractsVersion:     cfg.BassetContractsVersion,
		HubContract:                cfg.Addresses.HubContract,
		ValidatorsRegistryContract: cfg.Addresses.ValidatorsRegistryContract,
		UseActiveSetValidatorsList: true, // TODO make envconfig
	}
	validatorsRepository, err := repositories.NewValidatorsRepository(valRepoCfg, c.apiClient)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise a validators repository: %v", err)
	}
	delegatorsRepository := delegations.New(c.apiClient)
	signInfoRepository := signinfo.New(c.apiClient)

	slashingMonitor := monitors.NewSlashingMonitor(cfg, logger, validatorsRepository, signInfoRepository)
	c.RegisterMonitor(ctx, cfg, slashingMonitor)

	updateGlobalIndexMonitor := monitors.NewUpdateGlobalIndexMonitor(cfg, logger)
	c.RegisterMonitor(ctx, cfg, updateGlobalIndexMonitor)

	hubParameters := monitors.NewHubParametersMonitor(cfg, logger)
	c.RegisterMonitor(ctx, cfg, &hubParameters)

	delegationsDistributionMonitor := monitors.NewDelegationsDistributionMonitor(cfg, logger, validatorsRepository,
		delegatorsRepository)
	c.RegisterMonitor(ctx, cfg, delegationsDistributionMonitor)

	configCRC32Monitor := monitors.NewConfigsCRC32Monitor(cfg, logger)
	c.RegisterMonitor(ctx, cfg, configCRC32Monitor)

	validatorsFeeMonitor := monitors.NewValidatorsFeeMonitor(cfg, logger, validatorsRepository)
	c.RegisterMonitor(ctx, cfg, validatorsFeeMonitor)

	oracleVotesMonitor := monitors.NewOracleVotesMonitor(cfg, logger, validatorsRepository)
	c.RegisterMonitor(ctx, cfg, oracleVotesMonitor)

	balanceMonitor := monitors.NewOperatorBotBalanceMonitor(cfg, logger)
	c.RegisterMonitor(ctx, cfg, balanceMonitor)

	failedRedelegationsMonitor := monitors.NewFailedRedelegationsMonitor(cfg, logger, validatorsRepository, delegatorsRepository)
	c.RegisterMonitor(ctx, cfg, failedRedelegationsMonitor)

	missedBlocksMonitor := monitors.NewMissedBlocksMonitor(cfg, logger, validatorsRepository)
	c.RegisterMonitor(ctx, cfg, missedBlocksMonitor)

	slashingParamsMonitor := monitors.NewSlashingParamsMonitor(cfg, logger)
	c.RegisterMonitor(ctx, cfg, slashingParamsMonitor)

	oracleParamsMonitor := monitors.NewOracleParamsMonitor(cfg, logger)
	c.RegisterMonitor(ctx, cfg, oracleParamsMonitor)

	repoConfig := repositories.ValidatorsRepositoryConfig{
		BAssetContractsVersion:     cfg.BassetContractsVersion,
		HubContract:                cfg.Addresses.HubContract,
		ValidatorsRegistryContract: cfg.Addresses.ValidatorsRegistryContract,
		UseActiveSetValidatorsList: false,
	}
	repo, err := repositories.NewValidatorsRepository(repoConfig, c.apiClient)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise a whitelisted validators repository: %v", err)
	}
	whitelistedValidatorsMonitor := monitors.NewWhitelistedValidatorsMonitor(cfg, logger, repo)
	c.RegisterMonitor(ctx, cfg, &whitelistedValidatorsMonitor)

	return c, nil
}

type Collector struct {
	Metrics       map[monitors.MetricName]monitors.Monitor
	MetricVectors map[monitors.MetricName]monitors.Monitor
	Monitors      []monitors.Monitor
	logger        *logrus.Logger
	apiClient     *client.TerraRESTApis
}

func (c Collector) GetApiClient() *client.TerraRESTApis {
	return c.apiClient
}

func (c Collector) GetLogger() *logrus.Logger {
	return c.logger
}

func (c Collector) ProvidedMetrics() []monitors.MetricName {
	var metrics []monitors.MetricName
	for m := range c.Metrics {
		metrics = append(metrics, m)
	}
	return metrics
}

func (c Collector) ProvidedMetricVectors() []monitors.MetricName {
	var metrics []monitors.MetricName
	for m := range c.MetricVectors {
		metrics = append(metrics, m)
	}
	return metrics
}

func (c Collector) Get(metric monitors.MetricName) (float64, error) {
	monitor, found := c.Metrics[metric]
	if !found {
		return 0, fmt.Errorf("monitor for metric \"%s\" not found", metric)
	}
	return monitor.GetMetrics()[metric].Get(), nil
}

func (c Collector) GetVector(metric monitors.MetricName) (*monitors.MetricVector, error) {
	monitor, found := c.MetricVectors[metric]
	if !found {
		return nil, fmt.Errorf("monitor for metric vector \"%s\" not found", metric)
	}
	return monitor.GetMetricVectors()[metric], nil
}

func findMaps(key monitors.MetricName, maps ...map[monitors.MetricName]monitors.Monitor) (monitors.Monitor, bool) {
	for _, m := range maps {
		if wantedMonitor, found := m[key]; found {
			return wantedMonitor, true
		}
	}
	return nil, false
}

func (c *Collector) RegisterMonitor(ctx context.Context, cfg config.CollectorConfig, m monitors.Monitor) {
	for metric := range m.GetMetrics() {
		if wantedMonitor, found := findMaps(metric, c.Metrics, c.MetricVectors); found {
			panic(fmt.Sprintf("register monitor %s failed. metrics collision. Monitor %s has declared metric %s", m.Name(), wantedMonitor.Name(), metric))
		}

		c.Metrics[metric] = m
	}
	for metric := range m.GetMetricVectors() {
		if wantedMonitor, found := findMaps(metric, c.Metrics, c.MetricVectors); found {
			panic(fmt.Sprintf("register monitor %s failed. metrics collision. Monitor %s has declared metric %s", m.Name(), wantedMonitor.Name(), metric))
		}

		c.MetricVectors[metric] = m
	}
	c.Monitors = append(c.Monitors, m)

	// first initial data fetching
	err := m.Handler(ctx)
	if err != nil {
		c.logger.Errorf("failed to update %s data: %+v\n", m.Name(), err)
	}

	// running fetching data in background
	tk := time.NewTicker(cfg.UpdateDataInterval)

	go monitors.MustRunMonitor(ctx, m, tk, c.logger)
}
