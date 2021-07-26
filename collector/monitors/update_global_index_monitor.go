package monitors

import (
	"context"
	"fmt"
	"strconv"

	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/lidofinance/terra-monitors/openapi/client/transactions"
	"github.com/lidofinance/terra-monitors/openapi/models"
	"github.com/sirupsen/logrus"
)

type UpdateGlobalIndexTxsVariants int

const (
	NonUpdateGlobalIndexTX UpdateGlobalIndexTxsVariants = iota
	SuccessfulUpdateGlobalIndexTX
	FailedUpdateGlobalIndex
)

const UpdateGlobalIndexBase64Encoded = "eyJ1cGRhdGVfZ2xvYmFsX2luZGV4Ijp7fX0="

const (
	UpdateGlobalIndexSuccessfulTxSinceLastCheck Metric = "update_global_index_successful_tx_since_last_check"
	UpdateGlobalIndexFailedTxSinceLastCheck     Metric = "update_global_index_failed_tx_since_last_check"
	UpdateGlobalIndexGasWanted                  Metric = "update_global_index_gas_wanted"
	UpdateGlobalIndexGasUsed                    Metric = "update_global_index_gas_used"
	UpdateGlobalIndexUUSDFee                    Metric = "update_global_index_uusd_fee"
)

const threshold int = 10

type UpdateGlobalIndexMonitor struct {
	metrics          map[Metric]float64
	ApiResponse      *models.GetTxListResult
	ContractAddress  string
	apiClient        *client.TerraLiteForTerra
	logger           *logrus.Logger
	lastMaxCheckedID int
}

func NewUpdateGlobalIndexMonitor(cfg config.CollectorConfig) UpdateGlobalIndexMonitor {
	m := UpdateGlobalIndexMonitor{
		metrics:         make(map[Metric]float64),
		ContractAddress: cfg.UpdateGlobalIndexBotAddress,
		apiClient:       cfg.GetTerraClient(),
		logger:          cfg.Logger,
	}

	return m
}

func (m UpdateGlobalIndexMonitor) Name() string {
	return "UpdateGlobalIndexMonitor"
}

func (m *UpdateGlobalIndexMonitor) InitMetrics() {
	m.metrics[UpdateGlobalIndexSuccessfulTxSinceLastCheck] = 0
	m.metrics[UpdateGlobalIndexGasWanted] = 0
	m.metrics[UpdateGlobalIndexGasUsed] = 0
	m.metrics[UpdateGlobalIndexUUSDFee] = 0
	m.metrics[UpdateGlobalIndexFailedTxSinceLastCheck] = 0
}

func (m *UpdateGlobalIndexMonitor) Handler(ctx context.Context) error {
	var offset *float64
	var fetchedTxs int
	var firstCheck bool
	if m.lastMaxCheckedID == 0 {
		firstCheck = true
	}

	iterations := 0
	var maxProcessedID int
	var alreadyProcessedFound bool
	m.InitMetrics()
	for iterations < threshold {
		p := transactions.GetV1TxsParams{}
		p.SetAccount(&m.ContractAddress)
		p.SetContext(ctx)
		p.SetOffset(offset)

		resp, err := m.apiClient.Transactions.GetV1Txs(&p)
		if err != nil {
			return fmt.Errorf("failed to fetch transaction history for UpdateGlobalIndexBotContract account: %w", err)
		}

		maxProcessedID, alreadyProcessedFound = m.processTransactions(resp.Payload.Txs, m.lastMaxCheckedID)
		fetchedTxs += len(resp.Payload.Txs)
		maxProcessedID = maxInt(m.lastMaxCheckedID, maxProcessedID)
		if alreadyProcessedFound || firstCheck {
			break
		}
		offset = resp.Payload.Next
		iterations++
	}
	m.lastMaxCheckedID = maxProcessedID
	if threshold == iterations {
		m.logger.Warning("update global index processing stopped due to requests threshold - ", threshold)
	}
	m.logger.Infoln("update global index txs fetched:", fetchedTxs)
	m.logger.Infoln("update global index state:", m.metrics)
	// m.updateMetrics()
	return nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *UpdateGlobalIndexMonitor) processTransactions(txs []*models.GetTxListResultTxs, previousMaxCheckedID int) (newMaxCheckedID int, alreadyProcessedFound bool) {
	for i, tx := range txs {
		if i == 0 {
			newMaxCheckedID = maxInt(int(*tx.ID), previousMaxCheckedID)
		}
		if previousMaxCheckedID == int(*tx.ID) {
			// we have already checked this and earlier transactions
			m.logger.Infoln("stopping processing, last checked transaction is found:", previousMaxCheckedID)
			alreadyProcessedFound = true
			break
		}
		switch isTxUpdateGlobalIndex(tx) {
		case SuccessfulUpdateGlobalIndexTX:
			m.metrics[UpdateGlobalIndexSuccessfulTxSinceLastCheck]++
		case FailedUpdateGlobalIndex:
			m.metrics[UpdateGlobalIndexFailedTxSinceLastCheck]++
			m.logger.Warning("failed tx detected: ", getTxRawLog(tx))
		case NonUpdateGlobalIndexTX:
		}
		m.metrics[UpdateGlobalIndexGasUsed] += gasUsed(m.logger, tx)
		m.metrics[UpdateGlobalIndexGasWanted] += gasWanted(m.logger, tx)
		m.metrics[UpdateGlobalIndexUUSDFee] += uusdFee(m.logger, tx)
	}
	return newMaxCheckedID, alreadyProcessedFound
}

func (m UpdateGlobalIndexMonitor) GetMetrics() map[Metric]float64 {
	return m.metrics
}

func getTxRawLog(tx *models.GetTxListResultTxs) string {
	if tx == nil || tx.RawLog == nil {
		return ""
	}
	return *tx.RawLog
}

func isTxUpdateGlobalIndex(tx *models.GetTxListResultTxs) UpdateGlobalIndexTxsVariants {
	if tx == nil || tx.Tx == nil || tx.Tx.Value == nil || len(tx.Tx.Value.Msg) == 0 {
		return NonUpdateGlobalIndexTX
	}
	for _, msg := range tx.Tx.Value.Msg {
		if msg.Value == nil || msg.Value.ExecuteMsg == nil {
			continue
		}
		if *msg.Value.ExecuteMsg == UpdateGlobalIndexBase64Encoded && len(tx.Logs) > 0 {
			return SuccessfulUpdateGlobalIndexTX
		} else if *msg.Value.ExecuteMsg == UpdateGlobalIndexBase64Encoded && len(tx.Logs) == 0 {
			return FailedUpdateGlobalIndex
		}
	}
	return NonUpdateGlobalIndexTX
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
