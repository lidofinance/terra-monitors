package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/lidofinance/terra-monitors/openapi/client/transactions"
	"github.com/lidofinance/terra-monitors/openapi/client/wasm"
	"strconv"
)

type ValidatorsRepository interface {
	GetValidatorsAddresses(ctx context.Context) ([]string, error)
	GetValidatorInfo(ctx context.Context, address string) (types.ValidatorInfo, error)
}

type V1ValidatorsRepository struct {
	hubContract string
	apiClient   *client.TerraLiteForTerra
}

func NewV1ValidatorsRepository(cfg config.CollectorConfig) *V1ValidatorsRepository {
	return &V1ValidatorsRepository{
		hubContract: cfg.HubContract,
		apiClient:   cfg.GetTerraClient(),
	}
}

func (r *V1ValidatorsRepository) GetValidatorsAddresses(ctx context.Context) ([]string, error) {
	hubReq, hubResp := types.GetHubWhitelistedValidatorsPair()

	reqRaw, err := json.Marshal(&hubReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal HubWhitelistedValidators request: %w", err)
	}

	p := wasm.GetWasmContractsContractAddressStoreParams{}
	p.SetContext(ctx)
	p.SetContractAddress(r.hubContract)
	p.SetQueryMsg(string(reqRaw))

	resp, err := r.apiClient.Wasm.GetWasmContractsContractAddressStore(&p)
	if err != nil {
		return nil, fmt.Errorf("failed to process HubWhitelistedValidators request: %w", err)
	}

	if err := resp.GetPayload().Validate(nil); err != nil {
		return nil, fmt.Errorf("failed to validate ValidatorsWhitelist: %w", err)
	}

	err = types.CastMapToStruct(resp.Payload.Result, &hubResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HubWhitelistedValidators body interface: %w", err)
	}

	return hubResp.Validators, nil
}

func (r *V1ValidatorsRepository) GetValidatorInfo(ctx context.Context, address string) (types.ValidatorInfo, error) {
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

	return types.ValidatorInfo{
		Address:        address,
		Moniker:        validatorInfoResponse.GetPayload().Result.Description.Moniker,
		PubKey:         *validatorInfoResponse.GetPayload().Result.ConsensusPubkey,
		CommissionRate: commissionRate,
	}, nil
}
