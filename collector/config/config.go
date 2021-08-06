package config

import (
	"github.com/lidofinance/terra-monitors/internal/logging"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/sirupsen/logrus"
)

const (
	DefaultHubContract                 = "terra1mtwph2juhj0rvjz7dy92gvl6xvukaxu8rfv8ts"
	DefaultRewardContract              = "terra17yap3mhph35pcwvhza38c2lkj7gzywzy05h7l0"
	DefaultBlunaTokenInfoContract      = "terra1kc87mu460fwkqte29rquh4hc20m54fxwtsx7gp"
	DefaultUpdateGlobalIndexBotAddress = "terra1eqpx4zr2vm9jwu2vas5rh6704f6zzglsayf2fy"

	// TODO: use an actual address after validators_registry deployment.
	DefaultStlunaTokenInfoContract  = "terra_dummy_stluna_contract"
	DefaultValidatorRegistryAddress = "terra_dummy_validators_registry"
	DefaultAirDropRegistryContract  = "terra_dummy_airdrop"
	DefaultRewardDispatcherContract = "terra_dummy_rewarddispatcher"
)

type CollectorConfig struct {
	Logger                      *logrus.Logger
	LCDEndpoint                 string
	HubContract                 string
	RewardContract              string
	BlunaTokenInfoContract      string
	StlunaTokenInfoContract     string
	UpdateGlobalIndexBotAddress string
	ValidatorRegistryAddress    string
	RewardDispatcherContract    string
	AirDropRegistryContract     string
	Schemes                     []string
}

func (c CollectorConfig) getSchemes() []string {
	if len(c.Schemes) > 0 {
		return c.Schemes
	}

	return []string{"https"}
}

func (c CollectorConfig) GetTerraClient() *client.TerraLiteForTerra {
	if c.LCDEndpoint == "" {
		return client.NewHTTPClient(nil)
	}

	transportConfig := &client.TransportConfig{
		Host:     c.LCDEndpoint,
		BasePath: "/",
		Schemes:  c.getSchemes(),
	}

	return client.NewHTTPClientWithConfig(nil, transportConfig)
}

func DefaultCollectorConfig() CollectorConfig {
	return CollectorConfig{
		Logger:                      logging.NewDefaultLogger(),
		HubContract:                 DefaultHubContract,
		RewardContract:              DefaultRewardContract,
		BlunaTokenInfoContract:      DefaultBlunaTokenInfoContract,
		UpdateGlobalIndexBotAddress: DefaultUpdateGlobalIndexBotAddress,

		// change the fields to appropriate contracts values
		ValidatorRegistryAddress: DefaultValidatorRegistryAddress,
		AirDropRegistryContract:  DefaultAirDropRegistryContract,
		RewardDispatcherContract: DefaultRewardDispatcherContract,
		StlunaTokenInfoContract:  DefaultStlunaTokenInfoContract,
	}
}
