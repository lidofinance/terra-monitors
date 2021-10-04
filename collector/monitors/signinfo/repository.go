package signinfo

import (
	"context"

	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/sirupsen/logrus"
)

type Repository interface {
	// Init - Initializes state of validator's signInfo
	// for columbus-4 implementation you have to pass a PublicKey as second argument
	// for columbus-5 implementation you have to pass a PublicKey's consensus address as second argument
	Init(ctx context.Context, PubKeyOrConsAddr string) error
	GetMissedBlockCounter() float64
	GetTombstoned() bool
}

func NewSignInfoRepository(cfg config.CollectorConfig, logger *logrus.Logger) Repository {
	switch cfg.NetworkGeneration {
	case config.NetworkGenerationColumbus4:
		return NewRepositoryCol4(cfg, logger)
	case config.NetworkGenerationColumbus5:
		return NewRepositoryCol5(cfg, logger)
	default:
		panic("unknown network generation. available variants: columbus-4 or columbus-5")
	}
}
