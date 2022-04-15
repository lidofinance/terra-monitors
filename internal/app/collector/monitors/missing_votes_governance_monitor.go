package monitors

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client"
	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client/governance"
	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/models"
	"github.com/lidofinance/terra-repositories/proposals"

	"github.com/lidofinance/terra-monitors/internal/app/config"
	"github.com/sirupsen/logrus"
)

var (
	AlertLimitMetric  MetricName = "missed_votes_limit"
	MissedVotesMetric MetricName = "missed_votes_count"
)

type MissingVotesGovernanceMonitor struct {
	apiClient           *client.TerraRESTApis
	repository          *proposals.Repository
	logger              *logrus.Logger
	lookbackLimit       int // M
	alertLimit          int // N
	monitoredValidators []string
	metricVectors       map[MetricName]*MetricVector
	lock                sync.RWMutex
}

func NewMissingVotesGovernanceMonitor(cfg config.CollectorConfig, logger *logrus.Logger, apiClient *client.TerraRESTApis) *MissingVotesGovernanceMonitor {
	proposalsRepository := proposals.New(apiClient)
	m := MissingVotesGovernanceMonitor{
		repository:          proposalsRepository,
		apiClient:           apiClient,
		logger:              logger,
		lookbackLimit:       cfg.MissingVotesGovernanceMonitor.LookbackLimit,
		alertLimit:          cfg.MissingVotesGovernanceMonitor.AlertLimit,
		monitoredValidators: cfg.MissingVotesGovernanceMonitor.MonitoredValidators,
		metricVectors:       make(map[MetricName]*MetricVector),
		lock:                sync.RWMutex{},
	}
	m.logger.Infof("Initialized last governance voted monitor with lookback of: %v, alertLimit: %v", m.lookbackLimit, m.alertLimit)
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
		m.logger.Errorf("Could not get proposalList. Error: %v", err)
		return err
	}

	// Aggregate not voted validators counts
	for _, item := range proposalList {
		proposalID, err := strconv.Atoi(*item.ID)
		if err != nil {
			m.logger.Errorf("Could not convert ID to integer. ID: %v", item.ID)
			return err
		}

		notVotedForProposal, err := m.FetchNotVotedValidators(ctx, proposalID)
		if err != nil {
			m.logger.Errorf("Could not get proposal for %v. Error: %v", item.ID, err)
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

func (m *MissingVotesGovernanceMonitor) providedMetricVectors() []MetricName {
	return []MetricName{MissedVotesMetric, AlertLimitMetric}
}

func (m *MissingVotesGovernanceMonitor) FetchLastProposals(ctx context.Context) ([]*models.GetProposalListResultProposals, error) {
	resp, err := m.apiClient.Governance.GetV1GovProposals(
		&governance.GetV1GovProposalsParams{
			Context: ctx,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposals with {%s}", err)
	}

	err = resp.GetPayload().Validate(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to validate proposals with {%s}", err)
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
	res := make([]string, 0)

	proposalVotes, err := m.repository.GetVotes(ctx, proposalID)
	if err != nil {
		return res, fmt.Errorf("failed to get votes with {%s}", err)
	}

	votedValidatorsSubset := make([]string, 0)
	for _, item := range proposalVotes {
		votedValidatorsSubset = append(votedValidatorsSubset, *item.Voter.AccountAddress)
	}

	shouldVote := m.monitoredValidators
	res = difference(shouldVote, votedValidatorsSubset)

	return res, nil
}

func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
