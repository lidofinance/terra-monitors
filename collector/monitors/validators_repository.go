package monitors

import (
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/internal/client"
	"github.com/sirupsen/logrus"
)

func NewValidatorsRepository(cfg config.CollectorConfig, logger *logrus.Logger) ValidatorsRepository {
	switch cfg.BassetContractsVersion {
	case config.V1Contracts:
		return &V1ValidatorsRepository{
			hubContract: cfg.Addresses.HubContract,
			apiClient:   client.New(cfg.LCD, logger),
		}
	case config.V2Contracts:
		return &V2ValidatorsRepository{
			validatorsRegistryContract: cfg.Addresses.ValidatorsRegistryContract,
			apiClient:                  client.New(cfg.LCD, logger),
		}
	default:
		panic("unknown contracts version")
	}
}
