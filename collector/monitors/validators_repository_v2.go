package monitors

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"strconv"

	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/lidofinance/terra-monitors/openapi/client/transactions"
	"github.com/lidofinance/terra-monitors/openapi/client/wasm"
)

type V2ValidatorsRepository struct {
	validatorsRegistryContract string
	apiClient                  *client.TerraLiteForTerra
}

func (r *V2ValidatorsRepository) GetValidatorsAddresses(ctx context.Context) ([]string, error) {
	valReq, valResp := types.GetValidatorRegistryValidatorsPair()

	reqRaw, err := json.Marshal(&valReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ValidatorRegistryValidatorsRequest request: %w", err)
	}

	p := wasm.GetWasmContractsContractAddressStoreParams{}
	p.SetContext(ctx)
	p.SetContractAddress(r.validatorsRegistryContract)
	p.SetQueryMsg(string(reqRaw))

	resp, err := r.apiClient.Wasm.GetWasmContractsContractAddressStore(&p)
	if err != nil {
		return nil, fmt.Errorf("failed to process ValidatorRegistryValidatorsRequest request: %w", err)
	}

	if err := resp.GetPayload().Validate(nil); err != nil {
		return nil, fmt.Errorf("failed to validate ValidatorsWhitelist: %w", err)
	}

	err = types.CastMapToStruct(resp.Payload.Result, &valResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ValidatorRegistryValidatorsResponse body interface: %w", err)
	}

	valAddresses := make([]string, len(valResp))
	for i, val := range valResp {
		valAddresses[i] = val.Address
	}
	return valAddresses, nil
}

func (r *V2ValidatorsRepository) GetValidatorInfo(ctx context.Context, address string) (types.ValidatorInfo, error) {
	validatorInfoResponse, err := r.apiClient.Transactions.GetStakingValidatorsValidatorAddr(
		&transactions.GetStakingValidatorsValidatorAddrParams{
			ValidatorAddr: address,
			Context:       ctx,
		},
	)
	if err != nil {
		return types.ValidatorInfo{}, fmt.Errorf("failed to GetStakingValidatorsValidatorAddr: %w", err)
	}

	if err := validatorInfoResponse.GetPayload().Validate(nil); err != nil {
		return types.ValidatorInfo{}, fmt.Errorf("failed to validate ValidatorInfo for validator %s: %w", address, err)
	}

	commissionRate, err := strconv.ParseFloat(*validatorInfoResponse.GetPayload().Result.Commission.CommissionRates.Rate, 64)
	if err != nil {
		return types.ValidatorInfo{}, fmt.Errorf("failed to parse validator's comission rate: %w", err)
	}

	key, err := base64.StdEncoding.DecodeString(validatorInfoResponse.GetPayload().Result.ConsensusPubkey.Value)
	if err != nil {
		return types.ValidatorInfo{}, fmt.Errorf("failed to decode validator's ConsensusPubkey: %w", err)
	}

	pub := &ed25519.PubKey{Key: key}

	conPubKeyAddress, err := bech32.ConvertAndEncode(Bech32TerraValConsPrefix, pub.Address())
	if err != nil {
		return types.ValidatorInfo{}, fmt.Errorf("failed to convert validator's ConsensusPubkeyAddress to bech32: %w", err)
	}

	return types.ValidatorInfo{
		Address:        address,
		Moniker:        validatorInfoResponse.GetPayload().Result.Description.Moniker,
		PubKey:         conPubKeyAddress,
		CommissionRate: commissionRate,
		Jailed:         validatorInfoResponse.GetPayload().Result.Jailed,
	}, nil
}
