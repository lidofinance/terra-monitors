package monitors

import (
	"context"
	"fmt"
	"strconv"

	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/lidofinance/terra-monitors/internal/logging"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/lidofinance/terra-monitors/openapi/client/transactions"
	"github.com/lidofinance/terra-monitors/openapi/models"
	"github.com/sirupsen/logrus"
)

var UpdateGlobalIndexBase64Encoded = "eyJ1cGRhdGVfZ2xvYmFsX2luZGV4Ijp7fX0="

var (
	UpdateGlobalIndexTxSinceLastCheck Metric = "update_global_index_tx_since_last_check"
	UpdateGlobaIndexGasWanted         Metric = "update_global_index_gas_wanted"
	UpdateGlobaIndexGasUsed           Metric = "update_global_index_gas_used"
	UpdateGlobaIndexUUSDFee           Metric = "update_global_index_uusd_fee"
)

type UpdateGlobalIndexMonitor struct {
	state            types.UpdateGlobalIndexBotState
	ApiResponse      *models.GetTxListResult
	ContractAddress  string
	apiClient        *client.TerraLiteForTerra
	logger           *logrus.Logger
	lastMaxCheckedID int
}

func NewUpdateGlobalIndexMonitor(address string, apiClient *client.TerraLiteForTerra, logger *logrus.Logger) UpdateGlobalIndexMonitor {
	m := UpdateGlobalIndexMonitor{
		state:           types.UpdateGlobalIndexBotState{},
		ContractAddress: address,
		apiClient:       apiClient,
		logger:          logger,
	}

	if apiClient == nil {
		m.apiClient = client.NewHTTPClient(nil)
	}
	if logger == nil {
		m.logger = logging.NewDefaultLogger()
	}
	return m
}

func (h UpdateGlobalIndexMonitor) Name() string {
	return "UpdateGlobalIndexMonitor"
}

func (m *UpdateGlobalIndexMonitor) Handler(ctx context.Context) error {

	m.state = types.UpdateGlobalIndexBotState{}
	var offset *float64
	var fetchedTxs int
	for {
		p := transactions.GetV1TxsParams{}
		p.SetAccount(&m.ContractAddress)
		p.SetContext(ctx)
		p.SetOffset(offset)

		resp, err := m.apiClient.Transactions.GetV1Txs(&p)
		if err != nil {
			return fmt.Errorf("failed to fetch transaction history for UpdateGlobalIndexBotContract account: %w", err)
		}

		procesedState, maxID, alreadyProcessedFound := m.processTransactions(resp.Payload.Txs, m.lastMaxCheckedID)
		fetchedTxs += len(resp.Payload.Txs)
		m.state.GasUsedSinceLastCheck += procesedState.GasUsedSinceLastCheck
		m.state.GasWantedSinceLastCheck += procesedState.GasWantedSinceLastCheck
		m.state.SuccessfulTxSinceLastCheck += procesedState.SuccessfulTxSinceLastCheck
		m.state.UUSDFeeSinceLastCheck += procesedState.UUSDFeeSinceLastCheck
		if alreadyProcessedFound || m.lastMaxCheckedID == 0 {
			m.lastMaxCheckedID = maxID
			break
		}
		offset = resp.Payload.Next
	}

	m.logger.Infoln("update global index txs processed:", fetchedTxs)
	m.logger.Infoln("update global index state:", m.state)
	return nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m UpdateGlobalIndexMonitor) processTransactions(txs []*models.GetTxListResultTxs, previousMaxCheckedID int) (state types.UpdateGlobalIndexBotState, newMaxCheckedID int, alreadyProcessedFound bool) {
	for i, tx := range txs {
		if i == 0 {
			newMaxCheckedID = maxInt(int(*tx.ID), previousMaxCheckedID)
		}
		if previousMaxCheckedID == int(*tx.ID) {
			// we have already checked this and earlier transactions
			m.logger.Infoln("stopping processing, lastchecked transaction is found:", previousMaxCheckedID)
			alreadyProcessedFound = true
			break
		}
		if isSuccessfulTxUpdateGlobalIndex(tx) {
			state.SuccessfulTxSinceLastCheck++
		}

		state.GasUsedSinceLastCheck += gasUsed(m.logger, tx)
		state.GasWantedSinceLastCheck += gasWanted(m.logger, tx)
		state.UUSDFeeSinceLastCheck += uusdFee(m.logger, tx)
	}
	return state, newMaxCheckedID, alreadyProcessedFound
}

func (m UpdateGlobalIndexMonitor) ProvidedMetrics() []Metric {
	return []Metric{
		UpdateGlobalIndexTxSinceLastCheck,
		UpdateGlobaIndexGasWanted,
		UpdateGlobaIndexGasUsed,
		UpdateGlobaIndexUUSDFee,
	}
}

func (m UpdateGlobalIndexMonitor) Get(metric Metric) (float64, error) {
	switch metric {
	case UpdateGlobalIndexTxSinceLastCheck:
		return m.state.SuccessfulTxSinceLastCheck, nil
	case UpdateGlobaIndexGasWanted:
		return m.state.GasWantedSinceLastCheck, nil
	case UpdateGlobaIndexGasUsed:
		return m.state.GasUsedSinceLastCheck, nil
	case UpdateGlobaIndexUUSDFee:
		return m.state.UUSDFeeSinceLastCheck, nil
	}
	return 0, &MetricDoesNotExistError{metricName: metric}
}

func isSuccessfulTxUpdateGlobalIndex(tx *models.GetTxListResultTxs) bool {
	if tx == nil || tx.Tx == nil || tx.Tx.Value == nil || len(tx.Tx.Value.Msg) == 0 {
		return false
	}
	for _, msg := range tx.Tx.Value.Msg {
		if msg.Value == nil || msg.Value.ExecuteMsg == nil {
			continue
		}
		if *msg.Value.ExecuteMsg == UpdateGlobalIndexBase64Encoded {
			return true
		}
	}
	return false
}

func gasUsed(logger *logrus.Logger, tx *models.GetTxListResultTxs) float64 {
	if tx == nil || tx.GasUsed == nil {
		return 0
	}

	gasUsed, err := strconv.ParseFloat(*tx.GasUsed, 64)
	if err != nil && logger != nil {
		logger.Errorln("failed to parse gasUsed:", err)
	}
	return gasUsed
}

func gasWanted(logger *logrus.Logger, tx *models.GetTxListResultTxs) float64 {
	if tx == nil || tx.GasWanted == nil {
		return 0
	}

	gasWanted, err := strconv.ParseFloat(*tx.GasWanted, 64)
	if err != nil && logger != nil {
		logger.Errorln("failed to parse gasWanted:", err)
	}
	return gasWanted
}

func uusdFee(logger *logrus.Logger, tx *models.GetTxListResultTxs) float64 {
	if tx == nil ||
		tx.Tx == nil ||
		tx.Tx.Value == nil ||
		tx.Tx.Value.Fee == nil ||
		len(tx.Tx.Value.Fee.Amount) == 0 {
		return 0
	}
	fee := 0.0
	for _, amount := range tx.Tx.Value.Fee.Amount {
		if amount.Denom == nil || amount.Amount == nil {
			if logger != nil {
				logger.Warningf(
					"incorrect amount or denom value. \"amount.Denom\"=%v, \"amount.Amount\" = %v \n",
					amount.Denom,
					amount.Amount,
				)
			}
			continue
		}
		if *amount.Denom == "uusd" {
			uusdFeeAmount, err := strconv.ParseFloat(*amount.Amount, 64)
			if err != nil && logger != nil {
				logger.Errorln("failed to parse uusdFeeAmount:", err)
				continue
			}
			fee += uusdFeeAmount
		} else {
			_, err := strconv.ParseFloat(*amount.Amount, 64)
			if err != nil && logger != nil {
				logger.Errorln("failed to parse unaccountedFee:", err)
				continue
			}
			if err == nil && logger != nil {
				logger.Warningf(
					"unaccountedFee in tx. \"amount.Denom\"=%s, \"amount.Amount\" = %s \n",
					*amount.Denom,
					*amount.Amount,
				)
			}
		}
	}
	return fee
}
