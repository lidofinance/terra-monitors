package config

import (
	"fmt"
	"time"

	"github.com/vrischmann/envconfig"
)

const (
	V1Contracts = "1"
	V2Contracts = "2"

	NetworkGenerationColumbus5 = "columbus-5"
)

type CollectorConfig struct {
	BassetContractsVersion        string `envconfig:"default=2"` // available values: 1 and 2
	Source                        Source
	Addresses                     Addresses
	UpdateDataInterval            time.Duration `envconfig:"default=30s"`
	DelegationsDistributionConfig DelegationsDistributionConfig
	NetworkGeneration             string `envconfig:"default=columbus-5"` // available values: columbus-5
	MissingVotesGovernanceMonitor MissingVotesGovernanceMonitor
}

func NewCollectorConfig() (CollectorConfig, error) {
	config := CollectorConfig{}
	if err := envconfig.Init(&config); err != nil {
		return config, fmt.Errorf("failed to init config: %w", err)
	}

	return config, nil
}

type Source struct {
	Endpoints []string `envconfig:"default=fcd.terra.dev"`
	Schemes   []string `envconfig:"default=https"`
}

type Addresses struct {
	HubContract                 string `envconfig:"default=terra1mtwph2juhj0rvjz7dy92gvl6xvukaxu8rfv8ts"`
	RewardContract              string `envconfig:"default=terra17yap3mhph35pcwvhza38c2lkj7gzywzy05h7l0"`
	BlunaTokenInfoContract      string `envconfig:"default=terra1kc87mu460fwkqte29rquh4hc20m54fxwtsx7gp"`
	ValidatorsRegistryContract  string `envconfig:"default=terra_dummy_validators_registry"` // TODO: actualize.
	RewardsDispatcherContract   string `envconfig:"default=terra_dummy_rewards_dispatcher"`  // TODO: actualize.
	AirDropRegistryContract     string `envconfig:"default=terra_dummy_airdrop"`             // TODO: actualize.
	UpdateGlobalIndexBotAddress string `envconfig:"default=terra1eqpx4zr2vm9jwu2vas5rh6704f6zzglsayf2fy"`
}

type DelegationsDistributionConfig struct {
	NumMedianAbsoluteDeviations int64 `envconfig:"default=3"`
}

type MissingVotesGovernanceMonitor struct {
	LookbackLimit       int      `envconfig:"default=30"`
	AlertLimit          int      `envconfig:"default=5"`
	MonitoredValidators []string `envconfig:"default=terra1v5hrqlv8dqgzvy0pwzqzg0gxy899rm4kdn0jp4,terra123gn6j23lmexu0qx5qhmgxgunmjcqsx8g5ueq2,terra15zcjduavxc5mkp8qcqs9eyhwlqwdlrzy6anwpg,terra1v5hrqlv8dqgzvy0pwzqzg0gxy899rm4kdn0jp4,terra1kprce6kc08a6l03gzzh99hfpazfjeczfpd6td0,terra1c9ye54e3pzwm3e0zpdlel6pnavrj9qqvqgf7ps,terra144l7c3uph5a7h62xd8u5et3rqvj3dqtvve3he0,terra1542ek7muegmm806akl0lam5vlqlph7spfs99vq,terra1sym8gyehrdsm03vdc44rg9sflg8zeuqwfd3384,terra1khfcg09plqw84jxy5e7fj6ag4s2r9wqsg5jt4x,terra15urq2dtp9qce4fyc85m6upwm9xul30496lytpd,terra1alpf6snw2d76kkwjv3dp4l7pcl6cn9uytq89zk,terra1nwrksgv2vuadma8ygs8rhwffu2ygk4j24pxxx0,terra175hhkyxmkp8hf2zrzka7cnn7lk6mudtv4nsp2x,terra13g7z3qq6f00qww3u4mpcs3xw5jhqwraswv3q3t,terra1jkqr2vfg4krfd4zwmsf7elfj07cjuzss3qsmhm,terra15cupwhpnxhgylxa8n4ufyvux05xu864jcrrkqa"`
}
