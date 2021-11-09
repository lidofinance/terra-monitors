package delegations

import (
	"context"
	"fmt"

	"github.com/lidofinance/terra-monitors/internal/pkg/client"

	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/sirupsen/logrus"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/go-openapi/strfmt"
	terraClient "github.com/lidofinance/terra-monitors/openapi/client_bombay"
	"github.com/lidofinance/terra-monitors/openapi/client_bombay/query"
)

type Delegation struct {
	query.DelegatorDelegationsOKBodyDelegationResponsesItems0Delegation
	DelegationAmount types.Int
}

type Repository interface {
	GetDelegationsFromAddress(ctx context.Context, address string) (ret []Delegation, err error)
}

func New(cfg config.CollectorConfig, logger *logrus.Logger) Repository {
	return &BaseRepository{apiClient: client.NewBombay(cfg.LCD, logger)}
}

type BaseRepository struct {
	apiClient *terraClient.TerraLiteForTerra
}

func (r *BaseRepository) GetDelegationsFromAddress(ctx context.Context, address string) (ret []Delegation, err error) {
	var paginationKey strfmt.Base64
	for {
		delegationsResponse, err := r.apiClient.Query.DelegatorDelegations(&query.DelegatorDelegationsParams{
			PaginationKey: &paginationKey,
			DelegatorAddr: address,
			Context:       ctx,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get whitelistedValidators of delegator: %w", err)
		}

		if err := delegationsResponse.GetPayload().Validate(nil); err != nil {
			return nil, fmt.Errorf("failed to validate delegator's validators response: %w", err)
		}

		if delegationsResponse.Payload == nil {
			return nil, fmt.Errorf("failed to validate delegator's validators response: %w", err)
		}

		paginationKey = nil
		if delegationsResponse.Payload.Pagination != nil {
			paginationKey = delegationsResponse.Payload.Pagination.NextKey
		}

		for _, response := range delegationsResponse.GetPayload().DelegationResponses {
			if err := response.Validate(nil); err != nil {
				return nil, fmt.Errorf("failed to validate response: %w", err)
			}

			if response.Delegation == nil {
				return nil, fmt.Errorf("failed to validate response: delegaion is nil")
			}

			delegatedAmount, ok := types.NewInt(0), false
			if response.Balance != nil {
				if delegatedAmount, ok = types.NewIntFromString(response.Balance.Amount); !ok {
					return nil, fmt.Errorf("failed to parse delegation balance amount: %w", err)
				}

				ret = append(ret, Delegation{
					DelegatorDelegationsOKBodyDelegationResponsesItems0Delegation: *response.Delegation,
					DelegationAmount: delegatedAmount})
			} else {
				return nil, fmt.Errorf("failed to get response balance: balance is nil")
			}
		}

		if len(paginationKey) == 0 {
			break
		}
	}

	return
}
