package main

import (
	"context"
	"flag"
	"net/http"

	"github.com/lidofinance/terra-monitors/internal/app"
	"github.com/lidofinance/terra-monitors/internal/app/collector"
	"github.com/lidofinance/terra-monitors/internal/app/collector/monitors"
	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories/delegations"
	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories/signinfo"
	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories/validators"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/app/extractor"
	"github.com/lidofinance/terra-monitors/internal/pkg/logging"
	"github.com/sirupsen/logrus"
)

var addr = flag.String("listen-address", ":8080",
	"The address to listen on for HTTP requests.")

func createCollector(cfg config.CollectorConfig, logger *logrus.Logger) (collector.LCDCollector, error) {
	ctx := context.Background()

	c := collector.NewLCDCollector(cfg, logger)

	hubStateMonitor := monitors.NewHubStateMonitor(cfg, logger)
	c.RegisterMonitor(ctx, cfg, hubStateMonitor)

	rewardStateMonitor := monitors.NewRewardStateMonitor(cfg, logger)
	c.RegisterMonitor(ctx, cfg, &rewardStateMonitor)

	blunaTokenInfoMonitor := monitors.NewBlunaTokenInfoMonitor(cfg, logger)
	c.RegisterMonitor(ctx, cfg, blunaTokenInfoMonitor)

	validatorsRepository := validators.NewValidatorsRepository(cfg, c.GetLogger())
	delegatorsRepository := delegations.New(cfg, c.GetLogger())

	signInfoRepository := signinfo.NewSignInfoRepository(cfg, c.GetLogger())
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

	whitelistedValidatorsMonitor := monitors.NewWhitelistedValidatorsMonitor(cfg, logger, validatorsRepository)
	c.RegisterMonitor(ctx, cfg, &whitelistedValidatorsMonitor)

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

	return c, nil
}

func main() {
	flag.Parse()

	logger := logging.NewDefaultLogger()

	cfg, err := config.NewCollectorConfig()
	if err != nil {
		logger.Fatalf("Failed to create NewCollectorConfig: %s", err)
	}

	col, err := createCollector(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to createCollector(): %s", err)
	}

	var (
		promExtractor = extractor.NewPromExtractor(&col, logger)
		appInstance   = app.NewAppHTTP(promExtractor)
	)
	http.Handle("/metrics", appInstance)
	logger.Printf("Starting web server v%s at %s\n", cfg.BassetContractsVersion, *addr)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		logger.Errorf("Failed to ListenAndServe: %v\n", err)
	}
}
