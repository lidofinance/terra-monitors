package monitors

import "github.com/lidofinance/terra-monitors/collector/config"

func NewValidatorsRepository(cfg config.CollectorConfig) ValidatorsRepository {
	switch cfg.BassetContractsVersion {
	case config.V1Contracts:
		return &V1ValidatorsRepository{
			hubContract: cfg.Addresses.HubContract,
			apiClient:   cfg.GetTerraClient(),
		}
	case config.V2Contracts:
		return &V2ValidatorsRepository{
			validatorsRegistryContract: cfg.Addresses.ValidatorsRegistryContract,
			apiClient:                  cfg.GetTerraClient(),
		}
	default:
		panic("unknown contracts version")
	}
}
