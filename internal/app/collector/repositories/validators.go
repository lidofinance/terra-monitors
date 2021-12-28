package repositories

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/lidofinance/terra-monitors/internal/app/config"

	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client"
	"github.com/lidofinance/terra-repositories/validators"

	"github.com/cosmos/cosmos-sdk/types/bech32"
)

const (
	// BAssetContractsVersion1 refers to Anchor developer contracts (bLuna).
	BAssetContractsVersion1 = "1"
	// BAssetContractsVersion2 refers to contracts developed by Anchor and Lido (stLuna).
	BAssetContractsVersion2 = "2"
)

type ValidatorsRepositoryConfig struct {
	// BAssetContractsVersion is the generation of contracts to be used by the repository.
	BAssetContractsVersion string
	// HubContract refers to the contract containing information about whitelisted validators for
	// the 1st version of the bAsset contracts.
	HubContract string
	// ValidatorsRegistryContract refers to the contract containing information about whitelisted
	// validators for the 2nd version of the bAsset contracts.
	ValidatorsRegistryContract string
}

func NewValidatorsRepository(cfg ValidatorsRepositoryConfig, apiClient *client.TerraRESTApis) (ValidatorsRepository, error) {
	switch cfg.BAssetContractsVersion {
	case BAssetContractsVersion1:
		return validators.NewV1Repository(cfg.HubContract, apiClient), nil
	case BAssetContractsVersion2:
		return validators.NewV2Repository(cfg.ValidatorsRegistryContract, apiClient), nil
	default:
		return nil, fmt.Errorf("invalid bAsset contracts version %s", cfg.BAssetContractsVersion)
	}
}

type ValidatorsRepository interface {
	GetValidatorsAddresses(ctx context.Context) ([]string, error)
	GetValidatorInfo(ctx context.Context, address string) (validators.ValidatorInfo, error)
}

// GetValConsAddr - get valconsaddr from pubkeyidentifier.
// For columbus-5, pubkeyidentifier is valcons.
func GetValConsAddr(networkGeneration string, pubkeyidentifier string) (string, error) {
	var valconsAddr []byte
	var err error
	switch networkGeneration {
	case config.NetworkGenerationColumbus5:
		_, valconsAddr, err = bech32.DecodeAndConvert(pubkeyidentifier)
		if err != nil {
			return "", fmt.Errorf("failed to convert valcons(%s) to valconsaddr: %w", pubkeyidentifier, err)
		}
	default:
		panic("unknown network generation. available variants: columbus-5")
	}
	return strings.ToUpper(hex.EncodeToString(valconsAddr)), nil
}
