package monitors

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/lidofinance/terra-monitors/internal/client"
	"github.com/sirupsen/logrus"
)

func NewValidatorsRepository(cfg config.CollectorConfig, logger *logrus.Logger) ValidatorsRepository {
	switch cfg.BassetContractsVersion {
	case config.V1Contracts:
		return &V1ValidatorsRepository{
			hubContract:       cfg.Addresses.HubContract,
			apiClient:         client.New(cfg.LCD, logger),
			networkGeneration: cfg.NetworkGeneration,
		}
	case config.V2Contracts:
		return &V2ValidatorsRepository{
			validatorsRegistryContract: cfg.Addresses.ValidatorsRegistryContract,
			apiClient:                  client.New(cfg.LCD, logger),
			networkGeneration:          cfg.NetworkGeneration,
		}
	default:
		panic("unknown contracts version")
	}
}

type ValidatorsRepository interface {
	GetValidatorsAddresses(ctx context.Context) ([]string, error)
	GetValidatorInfo(ctx context.Context, address string) (types.ValidatorInfo, error)
}

func GetPubKeyIdentifier(networkGeneration string, pubkey interface{}) (string, error) {
	switch networkGeneration {
	case config.NetworkGenerationColumbus4:
		// columbus4 ConsensusPubkey is just a string
		pk, ok := pubkey.(string)
		if !ok {
			return "", fmt.Errorf("failed to cast pubkey interface to string: %+v", pubkey)
		}
		return pk, nil
	case config.NetworkGenerationColumbus5:
		// columbus5 ConsensusPubkey is a struct
		// "consensus_pubkey": {
		//      "type": "tendermint/PubKeyEd25519",
		//      "value": "EAI7kGuMo6BG1poseFcoMiSa4vHmXcYM4VCpFeIMncw="
		//    }
		pk, ok := pubkey.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("failed to cast pubkey interface to map[string]string: %+v", pubkey)
		}
		pubkeyValue, ok := pk["value"].(string)
		if !ok {
			return "", fmt.Errorf("failed to get pubkey's value from data struct: %+v", pk)
		}
		key, err := base64.StdEncoding.DecodeString(pubkeyValue)
		if err != nil {
			return "", fmt.Errorf("failed to decode validator's ConsensusPubkey: %w", err)
		}

		pub := &ed25519.PubKey{Key: key}

		consPubKeyAddress, err := bech32.ConvertAndEncode(Bech32TerraValConsPrefix, pub.Address())
		if err != nil {
			return "", fmt.Errorf("failed to convert validator's ConsensusPubkeyAddress to bech32: %w", err)
		}
		return consPubKeyAddress, nil

	default:
		panic("unknown network generation. available variants: columbus-4 or columbus-5")
	}
}
