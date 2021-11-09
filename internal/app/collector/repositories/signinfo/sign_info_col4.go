package signinfo

import (
	"context"
	"fmt"
	"strconv"

	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/client"
	terraClient "github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/lidofinance/terra-monitors/openapi/client/transactions"
	"github.com/lidofinance/terra-monitors/openapi/models"
	"github.com/sirupsen/logrus"
)

func NewRepositoryCol4(cfg config.CollectorConfig, logger *logrus.Logger) *RepositoryColumbus4 {
	return &RepositoryColumbus4{
		apiClient: client.New(cfg.LCD, logger),
		logger:    logger,
	}
}

type RepositoryColumbus4 struct {
	apiClient   *terraClient.TerraLiteForTerra
	logger      *logrus.Logger
	signingInfo *models.SigningInfo
}

func (s *RepositoryColumbus4) Init(ctx context.Context, pubKey string) error {
	signingInfoResponse, err := s.apiClient.Transactions.GetSlashingValidatorsValidatorPubKeySigningInfo(
		&transactions.GetSlashingValidatorsValidatorPubKeySigningInfoParams{
			ValidatorPubKey: pubKey,
			Context:         ctx,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to GetSlashingSigningInfos for validator's pubkey %s: %w", pubKey, err)
	}
	if err := signingInfoResponse.GetPayload().Validate(nil); err != nil {
		return fmt.Errorf("failed to validate SignInfo for validator %s: %w", pubKey, err)
	}
	s.signingInfo = signingInfoResponse.GetPayload().Result
	return nil
}

func (s *RepositoryColumbus4) GetMissedBlockCounter() float64 {
	if s.signingInfo != nil {
		// No blocks is sent as "", not as "0".
		if len(*s.signingInfo.MissedBlocksCounter) > 0 {
			// If the current block is greater than minHeight and the validator's MissedBlocksCounter is
			// greater than maxMissed, they will be slashed. So numMissedBlocks > 0 does not mean that we
			// are already slashed, but is alarming. Note: Liveness slashes do NOT lead to a tombstoning.
			// https://docs.terra.money/dev/spec-slashing.html#begin-block
			numMissedBlocks, err := strconv.ParseInt(*s.signingInfo.MissedBlocksCounter, 10, 64)
			if err != nil {
				s.logger.Errorf("failed to Parse `missed_blocks_counter:`: %s", err)
			} else {
				if numMissedBlocks > 0 {
					return float64(numMissedBlocks)
				}
			}
		}
	}
	return 0
}

func (s *RepositoryColumbus4) GetTombstoned() bool {
	if s.signingInfo != nil {
		if s.signingInfo.Tombstoned != nil {
			return *s.signingInfo.Tombstoned
		}
	}
	return false
}

func (s *RepositoryColumbus4) GetAddress() string {
	if s.signingInfo != nil {
		if s.signingInfo.Address != nil {
			return *s.signingInfo.Address
		}
	}
	return ""
}
