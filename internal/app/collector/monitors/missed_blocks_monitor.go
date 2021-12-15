package monitors

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/lidofinance/terra-monitors/internal/app/collector/repositories"
	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"

	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client"
	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client/tendermint_rpc"
	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/models"

	"github.com/sirupsen/logrus"
)

const (
	MissedBlocksForPeriod MetricName = "missed_blocks_for_period"
	InitialBlocksAmount              = 10
)

type MissedBlocksMonitor struct {
	networkGeneration      string
	validators             map[string]string // map valoper address -> valcons address
	latestCommittedChecked int
	metricVectors          map[MetricName]*MetricVector
	apiClient              *client.TerraRESTApis
	validatorsRepository   repositories.ValidatorsRepository
	logger                 *logrus.Logger
	lock                   sync.RWMutex
}

func NewMissedBlocksMonitor(
	cfg config.CollectorConfig,
	logger *logrus.Logger,
	repository repositories.ValidatorsRepository,
) *MissedBlocksMonitor {
	m := &MissedBlocksMonitor{
		networkGeneration:    cfg.NetworkGeneration,
		validators:           make(map[string]string),
		metricVectors:        make(map[MetricName]*MetricVector),
		apiClient:            utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger),
		validatorsRepository: repository,
		logger:               logger,
		lock:                 sync.RWMutex{},
	}

	m.InitMetrics()

	return m
}

func (m *MissedBlocksMonitor) Name() string {
	return "MissedBlocks"
}

func (m *MissedBlocksMonitor) providedMetrics() []MetricName {
	return []MetricName{}
}

func (m *MissedBlocksMonitor) providedMetricVectors() []MetricName {
	return []MetricName{
		MissedBlocksForPeriod,
	}
}

func (m *MissedBlocksMonitor) InitMetrics() {
	initMetrics(nil, m.providedMetricVectors(), nil, m.metricVectors)
}

func GetValidatorsSignedTheBlock(block *models.BlockQuery) map[string]struct{} {
	addresses := make(map[string]struct{})
	for _, signature := range block.Block.LastCommit.Signatures {
		addresses[signature.ValidatorAddress] = struct{}{}
	}
	return addresses
}

func (m *MissedBlocksMonitor) FetchLatestBlocks(ctx context.Context) ([]*models.BlockQuery, error) {
	var blocks []*models.BlockQuery
	var wg sync.WaitGroup
	var lock sync.Mutex

	req := tendermint_rpc.GetBlocksLatestParams{}
	req.SetContext(ctx)

	resp, err := m.apiClient.TendermintRPC.GetBlocksLatest(&req)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block info: %w", err)
	}

	if err := resp.GetPayload().Validate(nil); err != nil {
		return nil, fmt.Errorf("failed to validate latest block response: %w", err)
	}
	// last committed = 'height' - 1
	lastCommitted, err := strconv.Atoi(resp.GetPayload().Block.LastCommit.Height)
	if err != nil {
		return nil, fmt.Errorf("failed to parse commits height: %w", err)
	}

	// no new blocks
	if lastCommitted == m.latestCommittedChecked {
		return nil, nil
	}

	blocks = append(blocks, resp.GetPayload())

	if m.latestCommittedChecked == 0 {
		m.latestCommittedChecked = lastCommitted - InitialBlocksAmount
	}

	//fetching needed blocks to check signatures
	for committed := m.latestCommittedChecked + 1; committed < lastCommitted; committed++ {
		wg.Add(1)
		go func(committed int) {
			defer wg.Done()
			req := tendermint_rpc.GetBlocksHeightParams{}
			req.SetContext(ctx)
			// fetching block with height = committed + 1
			// we are checking signatures for committed block "height - 1" witch number in Block.LastCommit.Height field
			height := committed + 1
			req.SetHeight(int64(height))

			resp, err := m.apiClient.TendermintRPC.GetBlocksHeight(&req)
			if err != nil {
				m.logger.Errorf("failed to get latest block info: %+v\n", err)
				return
			}

			if err := resp.GetPayload().Validate(nil); err != nil {
				m.logger.Errorf("failed to validate latest block response: %+v", err)
				return
			}
			lock.Lock()
			defer lock.Unlock()
			blocks = append(blocks, resp.GetPayload())
		}(committed)
	}
	wg.Wait()
	m.latestCommittedChecked = lastCommitted
	return blocks, nil
}

func (m *MissedBlocksMonitor) Handler(ctx context.Context) error {
	// tmp* for 2stage nonblocking update data
	tmpMetricVectors := make(map[MetricName]*MetricVector)
	initMetrics(nil, []MetricName{MissedBlocksForPeriod}, nil, tmpMetricVectors)

	blocks, err := m.FetchLatestBlocks(ctx)

	if err != nil {
		return fmt.Errorf("failed to fetch blocks: %w", err)
	}
	if len(blocks) == 0 {
		m.logger.Infoln("no new blocks")
		return nil
	}

	validatorsInfo, err := getValidatorsInfo(ctx, m.validatorsRepository)

	if err != nil {
		return fmt.Errorf("failed to getValidatorsInfo: %w", err)
	}

	for _, validatorInfo := range validatorsInfo {

		for _, block := range blocks {
			consAddress, found := m.validators[validatorInfo.Address]
			if !found {
				consAddress, err = repositories.GetValConsAddr(m.networkGeneration, validatorInfo.PubKey)
				if err != nil {
					m.logger.Errorf("failed to convert pubkey identifier(%s) to addr : %+v", validatorInfo.PubKey, err)
					continue
				}
				m.validators[validatorInfo.Address] = consAddress
			}
			signedValidators := GetValidatorsSignedTheBlock(block)
			tmpMetricVectors[MissedBlocksForPeriod].Add(validatorInfo.Moniker, 0)
			if _, found := signedValidators[consAddress]; !found {
				tmpMetricVectors[MissedBlocksForPeriod].Add(validatorInfo.Moniker, 1)
			}
		}
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	// accumulating missed blocks
	for _, label := range tmpMetricVectors[MissedBlocksForPeriod].Labels() {
		m.metricVectors[MissedBlocksForPeriod].Add(label, tmpMetricVectors[MissedBlocksForPeriod].Get(label))
	}

	m.logger.Infoln("updated", m.Name())
	return nil
}

func (m *MissedBlocksMonitor) GetMetrics() map[MetricName]MetricValue {
	return nil
}

func (m *MissedBlocksMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	// we need a delta value between checks, so we are dropping missed blocks counter once value is read
	m.lock.Lock()
	defer m.lock.Unlock()
	r := make(map[MetricName]*MetricVector)
	copyVectors(m.metricVectors, r)
	initMetrics(nil, []MetricName{MissedBlocksForPeriod}, nil, m.metricVectors)
	return r
}
