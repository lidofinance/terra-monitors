package monitors

import (
	"context"
	"fmt"
	"sync"

	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"

	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client"
	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client/transactions"
	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/models"

	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"
)

type UpdateGlobalIndexTxsVariants int

const (
	NonUpdateGlobalIndexTX UpdateGlobalIndexTxsVariants = iota
	SuccessfulUpdateGlobalIndexTX
	FailedUpdateGlobalIndexTx
)

const UpdateGlobalIndexMsg = "update_global_index"
const UpdateGlobalIndexBase64Encoded = "eyJ1cGRhdGVfZ2xvYmFsX2luZGV4Ijp7fX0="

const (
	UpdateGlobalIndexSuccessfulTxSinceLastCheck MetricName = "update_global_index_successful_tx_since_last_check"
	UpdateGlobalIndexFailedTxSinceLastCheck     MetricName = "update_global_index_failed_tx_since_last_check"
	UpdateGlobalIndexGasWanted                  MetricName = "update_global_index_gas_wanted"
	UpdateGlobalIndexGasUsed                    MetricName = "update_global_index_gas_used"
	UpdateGlobalIndexUUSDFee                    MetricName = "update_global_index_uusd_fee"
)

const threshold int = 10

type UpdateGlobalIndexMonitor struct {
	ContractAddress   string
	metrics           map[MetricName]MetricValue
	apiClient         *client.TerraRESTApis
	logger            *logrus.Logger
	lastMaxCheckedID  int64
	lock              sync.RWMutex
	networkGeneration string
}

func NewUpdateGlobalIndexMonitor(cfg config.CollectorConfig, logger *logrus.Logger) *UpdateGlobalIndexMonitor {
	m := UpdateGlobalIndexMonitor{
		ContractAddress:   cfg.Addresses.UpdateGlobalIndexBotAddress,
		metrics:           make(map[MetricName]MetricValue),
		apiClient:         utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger),
		logger:            logger,
		lock:              sync.RWMutex{},
		networkGeneration: cfg.NetworkGeneration,
	}
	m.InitMetrics()

	return &m
}

func (m *UpdateGlobalIndexMonitor) Name() string {
	return "UpdateGlobalIndexMonitor"
}

func (m *UpdateGlobalIndexMonitor) providedMetrics() []MetricName {
	return []MetricName{
		UpdateGlobalIndexSuccessfulTxSinceLastCheck,
		UpdateGlobalIndexGasWanted,
		UpdateGlobalIndexGasUsed,
		UpdateGlobalIndexUUSDFee,
		UpdateGlobalIndexFailedTxSinceLastCheck,
	}
}

func (m *UpdateGlobalIndexMonitor) InitMetrics() {
	for _, metric := range m.providedMetrics() {
		if m.metrics[metric] == nil {
			m.metrics[metric] = &ReadOnceMetric{}
		}
		m.metrics[metric].Set(0)
	}
}

func (m *UpdateGlobalIndexMonitor) Handler(ctx context.Context) error {
	var offset *int64
	var fetchedTxs int
	var firstCheck bool
	if m.lastMaxCheckedID == 0 {
		firstCheck = true
	}

	iterations := 0
	var maxProcessedID int64
	var maxProcessedIDPerRequest int64
	var alreadyProcessedFound bool
	for iterations < threshold {
		p := transactions.GetV1TxsParams{}
		p.SetAccount(&m.ContractAddress)
		p.SetContext(ctx)
		p.SetOffset(offset)

		resp, err := m.apiClient.Transactions.GetV1Txs(&p)
		if err != nil {
			return fmt.Errorf("failed to fetch transaction history for UpdateGlobalIndexBotContract account: %w", err)
		}

		maxProcessedIDPerRequest, alreadyProcessedFound = m.processTransactions(resp.Payload.Txs, m.lastMaxCheckedID)
		fetchedTxs += len(resp.Payload.Txs)
		maxProcessedID = maxInt(maxProcessedID, maxProcessedIDPerRequest)
		if alreadyProcessedFound || firstCheck {
			break
		}
		offset = &resp.Payload.Next
		iterations++
	}
	m.lastMaxCheckedID = maxProcessedID
	if threshold == iterations {
		m.logger.Warning("update global index processing stopped due to requests threshold - ", threshold)
	}
	m.logger.Infoln("update global index txs fetched:", fetchedTxs)
	m.logger.Infoln("update global index state:", m.metrics)
	return nil
}

func maxInt(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func (m *UpdateGlobalIndexMonitor) processTransactions(
	txs []*models.GetTxListResultTxs,
	previousMaxCheckedID int64,
) (newMaxCheckedID int64, alreadyProcessedFound bool) {
	// transactions are reverse ordered by ID field
	for i, tx := range txs {
		if i == 0 {
			newMaxCheckedID = maxInt(tx.ID, previousMaxCheckedID)
		}
		if previousMaxCheckedID == tx.ID {
			// we have already checked this and earlier transactions
			m.logger.Infoln("stopping processing, last checked transaction is found:", previousMaxCheckedID)
			alreadyProcessedFound = true
			break
		}
		switch isTxUpdateGlobalIndex(tx, m.networkGeneration) {
		case SuccessfulUpdateGlobalIndexTX:
			m.metrics[UpdateGlobalIndexSuccessfulTxSinceLastCheck].Add(1)
		case FailedUpdateGlobalIndexTx:
			m.metrics[UpdateGlobalIndexFailedTxSinceLastCheck].Add(1)
			m.logger.Warning("failed tx detected: ", getTxRawLog(tx))
		case NonUpdateGlobalIndexTX:
		}
		m.metrics[UpdateGlobalIndexGasUsed].Add(gasUsed(m.logger, tx))
		m.metrics[UpdateGlobalIndexGasWanted].Add(gasWanted(m.logger, tx))
		m.metrics[UpdateGlobalIndexUUSDFee].Add(uusdFee(m.logger, tx))
	}
	return newMaxCheckedID, alreadyProcessedFound
}

func (m *UpdateGlobalIndexMonitor) GetMetrics() map[MetricName]MetricValue {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.metrics
}

func (m *UpdateGlobalIndexMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	return nil
}

func getTxRawLog(tx *models.GetTxListResultTxs) string {
	if tx == nil || tx.RawLog == nil {
		return ""
	}
	return *tx.RawLog
}

func isTxUpdateGlobalIndex(tx *models.GetTxListResultTxs, networkGeneration string) UpdateGlobalIndexTxsVariants {
	if tx == nil || tx.Tx == nil || tx.Tx.Value == nil || len(tx.Tx.Value.Msg) == 0 {
		return NonUpdateGlobalIndexTX
	}
	for _, msg := range tx.Tx.Value.Msg {
		if msg.Value == nil || msg.Value.ExecuteMsg == nil {
			continue
		}
		var isUpdateGlobalIndexMsg bool
		switch networkGeneration {
		case config.NetworkGenerationColumbus5:
			isUpdateGlobalIndexMsg = isUpdateGlobalIndexMsgColumbus5(msg)
		default:
			panic("unknown network generation. available variants: columbus-5")
		}
		if isUpdateGlobalIndexMsg {
			if len(tx.Logs) == 0 {
				// https://fcd.terra.dev/v1/txs?offset=126987824
				// tx with id = 126987823 is a failed tx due to out of gas
				// as we can see there are two signs of failed transaction. The first one - there is no "logs" field in json response.
				// The second one - "raw_log" contains human-readable message with error
				return FailedUpdateGlobalIndexTx
			}
			return SuccessfulUpdateGlobalIndexTX
		}
	}
	return NonUpdateGlobalIndexTX
}

func isUpdateGlobalIndexMsgColumbus5(msg *models.GetTxListResultTxsTxValueMsg) bool {
	// columbus-5 execute message format
	// "execute_msg": {
	//     "update_global_index": {}
	//  },
	m, ok := msg.Value.ExecuteMsg.(map[string]interface{})
	if !ok {
		return false
	}
	if _, found := m[UpdateGlobalIndexMsg]; found {
		return true
	}
	return false
}

func gasUsed(logger *logrus.Logger, tx *models.GetTxListResultTxs) float64 {
	if tx == nil || tx.GasUsed == nil {
		return 0
	}

	gasUsed, err := cosmostypes.NewDecFromStr(*tx.GasUsed)
	if err != nil && logger != nil {
		logger.Errorln("failed to parse gasUsed:", err)
	}

	gasUsedValue, err := gasUsed.Float64()
	if err != nil {
		logger.Errorln("failed to parse gasUsed:", err)
	}

	return gasUsedValue
}

func gasWanted(logger *logrus.Logger, tx *models.GetTxListResultTxs) float64 {
	if tx == nil || tx.GasWanted == nil {
		return 0
	}

	gasWanted, err := cosmostypes.NewDecFromStr(*tx.GasWanted)
	if err != nil && logger != nil {
		logger.Errorln("failed to parse gasWanted:", err)
	}

	gasWantedValue, err := gasWanted.Float64()
	if err != nil {
		logger.Errorln("failed to parse gasWanted:", err)
	}

	return gasWantedValue
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
		if *amount.Denom == UUSDDenom {
			uusdFeeAmount, err := cosmostypes.NewDecFromStr(*amount.Amount)
			if err != nil && logger != nil {
				logger.Errorln("failed to parse uusdFeeAmount:", err)
				continue
			}

			uusdFeeAmountValue, err := uusdFeeAmount.Float64()
			if err != nil {
				logger.Errorln("failed to parse uusdFeeAmount:", err)
				continue
			}

			fee += uusdFeeAmountValue
		} else {
			_, err := cosmostypes.NewDecFromStr(*amount.Amount)
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
