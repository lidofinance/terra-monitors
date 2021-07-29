package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/lidofinance/terra-monitors/openapi/client/wasm"

	"github.com/lidofinance/terra-monitors/openapi/client/transactions"

	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/sirupsen/logrus"
)

const (
	SlashingNumJailedValidators     Metric = "slashing_num_jailed_validators"
	SlashingNumTombstonedValidators Metric = "slashing_num_tombstoned_validators"
	SlashingNumMissedBlocks         Metric = "slashing_num_missed_blocks" // TODO: add a "validator_address" label.
)

const (
	jailTimeLayout = "2006-01-02T15:04:05Z"
)

type SlashingMonitor struct {
	metrics              map[Metric]float64
	apiClient            *client.TerraLiteForTerra
	validatorsRepository ValidatorsRepository
	logger               *logrus.Logger
}

func NewSlashingMonitor(cfg config.CollectorConfig, repository ValidatorsRepository) *SlashingMonitor {
	m := &SlashingMonitor{
		metrics:              make(map[Metric]float64),
		apiClient:            cfg.GetTerraClient(),
		validatorsRepository: repository,
		logger:               cfg.Logger,
	}

	return m
}

func (m *SlashingMonitor) Name() string {
	return "Slashing"
}

func (m *SlashingMonitor) InitMetrics() {
	m.metrics = map[Metric]float64{
		SlashingNumJailedValidators:     0,
		SlashingNumTombstonedValidators: 0,
		SlashingNumMissedBlocks:         0,
	}
}

func (m *SlashingMonitor) Handler(ctx context.Context) error {
	m.InitMetrics()

	validatorPublicKeys, err := m.getValidatorsPublicKeys(ctx)
	if err != nil {
		return fmt.Errorf("failed to getValidatorsPublicKeys: %w", err)
	}

	for _, validatorPublicKey := range validatorPublicKeys {
		signingInfoResponse, err := m.apiClient.Transactions.GetSlashingValidatorsValidatorPubKeySigningInfo(
			&transactions.GetSlashingValidatorsValidatorPubKeySigningInfoParams{
				ValidatorPubKey: validatorPublicKey,
				Context:         ctx,
			},
		)
		if err != nil {
			m.logger.Errorf("failed to GetSlashingSigningInfos for validator %s: %s", validatorPublicKey, err)
			continue
		}

		var signingInfo = signingInfoResponse.GetPayload().Result

		// This check is just in case, I haven't seen empty values in this
		// field.
		if len(*signingInfo.JailedUntil) > 0 {
			jailedUntil, err := time.Parse(jailTimeLayout, *signingInfo.JailedUntil)
			if err != nil {
				m.logger.Errorf("failed to Parse `jailed_until` %s as %s: %s",
					*signingInfo.JailedUntil, jailTimeLayout, err)
			} else {
				// If the `jailed_until` property is set to a date in the future, this
				// validator is jailed.
				if jailedUntil.After(time.Now()) {
					m.metrics[SlashingNumJailedValidators]++
				}
			}
		}

		// No blocks is sent as "", not as "0".
		if len(*signingInfo.MissedBlocksCounter) > 0 {
			// If the current block is greater than minHeight and the validator's MissedBlocksCounter is
			// greater than maxMissed, they will be slashed. So numMissedBlocks > 0 does not mean that we
			// are already slashed, but is alarming. Note: Liveness slashes do NOT lead to a tombstoning.
			// https://docs.terra.money/dev/spec-slashing.html#begin-block
			numMissedBlocks, err := strconv.ParseInt(*signingInfo.MissedBlocksCounter, 10, 64)
			if err != nil {
				m.logger.Errorf("failed to Parse `missed_blocks_counter:`: %s", err)
			} else {
				if numMissedBlocks > 0 {
					m.metrics[SlashingNumMissedBlocks] += float64(numMissedBlocks)
				}
			}
		}

		if *signingInfo.Tombstoned {
			m.metrics[SlashingNumTombstonedValidators]++
		}
	}

	return nil
}

func (m *SlashingMonitor) GetMetrics() map[Metric]float64 {
	return m.metrics
}

func (m *SlashingMonitor) getValidatorsPublicKeys(ctx context.Context) ([]string, error) {
	validatorsAddresses, err := m.validatorsRepository.GetValidatorsAddresses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to getWhitelistedValidatorsAddresses: %w", err)
	}

	// For each validator address, get the consensus public key (which is required to
	// later get the signing info).
	var validatorConsensusPublicKeys []string
	for _, validatorAddress := range validatorsAddresses {
		validatorInfoResponse, err := m.apiClient.Transactions.GetStakingValidatorsValidatorAddr(
			&transactions.GetStakingValidatorsValidatorAddrParams{
				ValidatorAddr: validatorAddress,
				Context:       ctx,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to GetStakingValidatorsValidatorAddr: %w", err)
		}

		validatorConsensusPublicKeys = append(validatorConsensusPublicKeys,
			*validatorInfoResponse.GetPayload().Result.ConsensusPubkey)
	}

	return validatorConsensusPublicKeys, nil
}

type ValidatorsRepository interface {
	GetValidatorsAddresses(ctx context.Context) ([]string, error)
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

	err = types.CastMapToStruct(resp.Payload.Result, &hubResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HubWhitelistedValidators body interface: %w", err)
	}

	return hubResp.Validators, nil
}
