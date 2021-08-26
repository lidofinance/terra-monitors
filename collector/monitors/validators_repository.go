package monitors

import "github.com/lidofinance/terra-monitors/collector/config"

func NewValidatorsRepository(cfg config.CollectorConfig) ValidatorsRepository {
	if cfg.BassetContractsVersion == "1" {
		return &V1ValidatorsRepository{
			hubContract: cfg.Addresses.HubContract,
			apiClient:   cfg.GetTerraClient(),
		}
	}
	return &V2ValidatorsRepository{
		validatorsRegistryContract: cfg.Addresses.ValidatorsRegistryContract,
		apiClient:                  cfg.GetTerraClient(),
	}
}
