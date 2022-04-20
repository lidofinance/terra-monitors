package monitors

import (
	"context"
	"fmt"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"
	"strconv"
	"sync"

	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client"
	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client/governance"
	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/models"
	"github.com/lidofinance/terra-repositories/proposals"
	"github.com/lidofinance/terra-repositories/validators"

	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/sirupsen/logrus"
)

var (
	AlertLimitMetric  MetricName = "missed_votes_limit"
	MissedVotesMetric MetricName = "missed_votes_count"
)

type MissingVotesGovernanceMonitor struct {
	apiClient            *client.TerraRESTApis
	proposalsRepository  *proposals.Repository
	validatorsRepository *validators.V2Repository
	logger               *logrus.Logger
	// lookbackLimit is how many proposals to look for in the past
	lookbackLimit int
	// alertLimit is how many votes can be missed before alerting
	alertLimit    int
	metricVectors map[MetricName]*MetricVector
	lock          sync.RWMutex
}

func NewMissingVotesGovernanceMonitor(cfg config.CollectorConfig, logger *logrus.Logger, apiClient *client.TerraRESTApis) *MissingVotesGovernanceMonitor {
	proposalsRepository := proposals.New(apiClient)
	validatorsRepository := validators.NewV2Repository(cfg.Addresses.ValidatorsRegistryContract, apiClient)
	m := MissingVotesGovernanceMonitor{
		proposalsRepository:  proposalsRepository,
		validatorsRepository: validatorsRepository,
		apiClient:            apiClient,
		logger:               logger,
		lookbackLimit:        cfg.MissingVotesGovernanceMonitor.LookbackLimit,
		alertLimit:           cfg.MissingVotesGovernanceMonitor.AlertLimit,
		metricVectors:        make(map[MetricName]*MetricVector),
		lock:                 sync.RWMutex{},
	}
	m.InitMetrics()

	return &m
}

func (m *MissingVotesGovernanceMonitor) Name() string {
	return "MissingVotesGovernanceMonitor"
}

func (m *MissingVotesGovernanceMonitor) GetMetrics() map[MetricName]MetricValue {
	return nil
}

func (m *MissingVotesGovernanceMonitor) GetMetricVectors() map[MetricName]*MetricVector {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.metricVectors
}

func (m *MissingVotesGovernanceMonitor) Handler(ctx context.Context) error {
	tmpMetricVectors := make(map[MetricName]*MetricVector)
	initMetrics(nil, m.providedMetricVectors(), nil, tmpMetricVectors)

	notVotedValidators := make(map[string]uint)

	proposalList, err := m.FetchLastProposals(ctx)
	if err != nil {
		m.logger.Errorf("could not get proposalList: %s", err)
		return err
	}

	// Aggregate not voted validators counts
	for _, item := range proposalList {
		proposalID, err := strconv.Atoi(*item.ID)
		if err != nil {
			m.logger.Errorf("could not convert ID %s to integer: ", *item.ID)
			return err
		}

		notVotedForProposal, err := m.FetchNotVotedValidators(ctx, proposalID)
		if err != nil {
			m.logger.Errorf("could not get proposal for %s: %s", *item.ID, err)
			return err
		}

		for _, validatorItem := range notVotedForProposal {
			notVotedValidators[validatorItem] += 1
		}
	}

	for address, notVotedCount := range notVotedValidators {
		tmpMetricVectors[MissedVotesMetric].Set(address, float64(notVotedCount))
		tmpMetricVectors[AlertLimitMetric].Set(address, float64(m.alertLimit))
	}

	m.logger.Infof("Successfully retrieved list of missing votes")

	m.lock.Lock()
	defer m.lock.Unlock()
	copyVectors(tmpMetricVectors, m.metricVectors)

	return nil
}

func (m *MissingVotesGovernanceMonitor) InitMetrics() {
	initMetrics(nil, m.providedMetricVectors(), nil, m.metricVectors)
}

func (m *MissingVotesGovernanceMonitor) FetchLastProposals(ctx context.Context) ([]*models.GetProposalListResultProposals, error) {
	resp, err := m.apiClient.Governance.GetV1GovProposals(
		&governance.GetV1GovProposalsParams{
			Context: ctx,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposals: %w", err)
	}

	err = resp.GetPayload().Validate(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to validate proposals: %w", err)
	}

	proposalsList := resp.GetPayload().Proposals

	result := make([]*models.GetProposalListResultProposals, 0)

	for _, item := range proposalsList {
		if *item.Status == "Passed" || *item.Status == "Rejected" {
			result = append(result, item)
		}

		if len(result) >= m.lookbackLimit {
			break
		}
	}

	return result, nil
}

func (m *MissingVotesGovernanceMonitor) FetchNotVotedValidators(ctx context.Context, proposalID int) ([]string, error) {
	proposalVotes, err := m.proposalsRepository.GetVotes(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get votes for %d: %w", proposalID, err)
	}

	votedValidators := make([]string, 0, len(proposalVotes))
	for _, item := range proposalVotes {
		votedValidators = append(votedValidators, *item.Voter.AccountAddress)
	}

	shouldVoteValidatorsValoper, err := m.validatorsRepository.GetValidatorsAddresses(ctx)
	shouldVoteValidators := make([]string, 0, len(shouldVoteValidatorsValoper))
	for _, item := range shouldVoteValidatorsValoper {
		address, err := utils.ValoperToAccAddress(item)
		if err != nil {
			return nil, fmt.Errorf("failed to convert valoperAddress to address for %s: %w", item, err)
		}
		shouldVoteValidators = append(shouldVoteValidators, address)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get lido validators from the contract: %w", err)
	}

	return utils.StringsSetsDifference(shouldVoteValidators, votedValidators), nil
}

func (m *MissingVotesGovernanceMonitor) providedMetricVectors() []MetricName {
	return []MetricName{MissedVotesMetric, AlertLimitMetric}
}
