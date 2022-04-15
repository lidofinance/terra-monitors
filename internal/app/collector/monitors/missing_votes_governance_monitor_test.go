package monitors

import (
	"context"
	"io/ioutil"

	"github.com/lidofinance/terra-monitors/internal/pkg/stubs"
	"github.com/lidofinance/terra-monitors/internal/pkg/utils"
	"github.com/stretchr/testify/suite"
)

type MissingVotesGovernanceMonitorTestSuite struct {
	suite.Suite
}

func (suite *MissingVotesGovernanceMonitorTestSuite) SetupTest() {
}

func (suite *MissingVotesGovernanceMonitorTestSuite) TestNoMissingVotesFound() {
	dir, err := utils.GetTerraMonitorsPath()
	suite.NoError(err)

	proposalsInfo, err := ioutil.ReadFile(dir + "test_data/governance_proposals_found.json")
	suite.NoError(err)
	proposalInfo, err := ioutil.ReadFile(dir + "test_data/governance_proposal_votes_found.json")
	suite.NoError(err)

	testServerResponses := map[string]string{
		"/v1/gov/proposals":           string(proposalsInfo),
		"/v1/gov/proposals/721/votes": string(proposalInfo),
		"/v1/gov/proposals/720/votes": string(proposalInfo),
		"/v1/gov/proposals/599/votes": string(proposalInfo),
	}

	testServer := stubs.NewServerWithRoutedResponse(testServerResponses)
	cfg := stubs.NewTestCollectorConfig(testServer.URL)
	logger := stubs.NewTestLogger()
	apiClient := utils.BuildClient(utils.SourceToEndpoints(cfg.Source), logger)
	m := NewMissingVotesGovernanceMonitor(cfg, logger, apiClient)

	err = m.Handler(context.Background())
	suite.NoError(err)

	votesMissed := m.metricVectors[MissedVotesMetric]
	votesLimits := m.metricVectors[AlertLimitMetric]

	missedAddress := "terra1ya8mnvt0c6rahzvgj4ezwkt49rsscn6lkeklul"
	missingVote := votesMissed.Get(missedAddress)
	missingLimit := votesLimits.Get(missedAddress)

	suite.Equal(3.0, missingVote)
	suite.Equal(10.0, missingLimit)
}
