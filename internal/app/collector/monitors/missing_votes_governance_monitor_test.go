package monitors

import (
	"context"
	"fmt"
	"github.com/lidofinance/terra-monitors/internal/app/collector/types"
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
	validators, err := ioutil.ReadFile(dir + "test_data/validators_registry_validators_response_2.json")
	suite.NoError(err)

	testServerResponses := map[string]string{
		"/v1/gov/proposals":           string(proposalsInfo),
		"/v1/gov/proposals/721/votes": string(proposalInfo),
		"/v1/gov/proposals/720/votes": string(proposalInfo),
		"/v1/gov/proposals/599/votes": string(proposalInfo),
		fmt.Sprintf("/wasm/contracts/%s/store", types.ValidatorsRegistryContract): string(validators),
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

	missedAddress := "terra15zcjduavxc5mkp8qcqs9eyhwlqwdlrzy6anwpg"
	missingVote := votesMissed.Get(missedAddress)
	missingLimit := votesLimits.Get(missedAddress)

	suite.Equal(3.0, missingVote)
	suite.Equal(10.0, missingLimit)
}
