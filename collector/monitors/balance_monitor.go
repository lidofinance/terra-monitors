package monitors

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/internal/client"
	terraClient "github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/lidofinance/terra-monitors/openapi/client/bank"
	"github.com/sirupsen/logrus"
)

var (
	OperatorBotBalance MetricName = "operator_bot_balance"
)

type OperatorBotBalanceMonitor struct {
	BotAddress string
	apiClient  *terraClient.TerraLiteForTerra
	logger     *logrus.Logger
	balanceUST SimpleMetricValue
}

func NewOperatorBotBalanceMonitor(cfg config.CollectorConfig, logger *logrus.Logger) *OperatorBotBalanceMonitor {
	m := OperatorBotBalanceMonitor{
		BotAddress: cfg.Addresses.UpdateGlobalIndexBotAddress,
		apiClient:  client.New(cfg.LCD, logger),
		logger:     logger,
		balanceUST: SimpleMetricValue{},
	}
	return &m
}

func (m *OperatorBotBalanceMonitor) Name() string {
	return "OperatorBotBalanceMonitor"
}

func (m *OperatorBotBalanceMonitor) GetMetrics() map[MetricName]MetricValue {
	return map[MetricName]MetricValue{
		OperatorBotBalance: &m.balanceUST,
	}
}

func (m *OperatorBotBalanceMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	return nil
}

func (m *OperatorBotBalanceMonitor) Handler(ctx context.Context) error {

	p := bank.GetBankBalancesAddressParams{}
	p.SetContext(ctx)
	p.SetAddress(m.BotAddress)

	resp, err := m.apiClient.Bank.GetBankBalancesAddress(&p)
	if err != nil {
		return fmt.Errorf("failed to get \"%s\" account balance: %w", m.BotAddress, err)
	}
	err = resp.GetPayload().Validate(nil)
	if err != nil {
		return fmt.Errorf("failed to validate response: %w", err)
	}
	coins := resp.GetPayload().Result
	for _, coin := range coins {
		if coin.Denom == UUSDDenom {
			amount, err := types.NewDecFromStr(coin.Amount)
			if err != nil {
				return fmt.Errorf("failed to parse coins uusd amount: %s", coin.Amount)
			}

			amountFloat, err := amount.Float64()
			if err != nil {
				return fmt.Errorf("failed to parse coins uusd amount: %s: %s", coin.Amount, err)
			}

			m.balanceUST.Set(float64(amountFloat / 1_000_000))
			m.logger.Infof("successfully retrieved \"%s\" account balance info\n", m.BotAddress)
			return nil
		}
	}
	// in case there is no uusd coins set balance to 0
	m.balanceUST.Set(0)
	return fmt.Errorf("uusd coin not found")
}
