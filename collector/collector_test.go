package collector

import (
	"bytes"
	"context"
	"testing"

	"github.com/lidofinance/terra-monitors/internal/logging"
	"github.com/stretchr/testify/suite"
)

type CollectorTestSuite struct {
	suite.Suite
	Collector LCDCollector
}

func (suite *CollectorTestSuite) SetupTest() {
	logger := logging.NewDefaultLogger()
	out := bytes.NewBuffer(nil)
	logger.Out = out
	suite.Collector = NewLCDCollector("http://dummy.com", "terra421h3krehwjkfls", logger)
	suite.Collector.HttpClient = NewMockClient()
}

func (suite *CollectorTestSuite) TestSuccessfullQueryRequest() {
	expected := TokenInfoResponse{
		Name:        "Bonded Luna",
		Symbol:      "BLUNA",
		Decimals:    6,
		TotalSupply: "79178685320809",
	}
	req, resp := GetCommonTokenInfoPair()
	err := suite.Collector.buildAndProcessRequest(context.Background(), suite.Collector.BlunaContractAddress, req, &resp)
	suite.Require().NoError(err)
	suite.Equal(expected, resp)
}

func (suite *CollectorTestSuite) TestBadQueryRequest() {
	resp := struct{}{}
	err := suite.Collector.buildAndProcessRequest(context.Background(), suite.Collector.BlunaContractAddress, struct{ ConnectionRefused string }{}, &resp)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "connection refused")
}

func (suite *CollectorTestSuite) TestConnectionRefusedRequest() {
	resp := struct {
		Error string `json:"error"`
	}{}
	req := struct {
		BadQuery struct{} `json:"bad_query"`
	}{}
	err := suite.Collector.buildAndProcessRequest(context.Background(), suite.Collector.BlunaContractAddress, req, &resp)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "parsing anchor_basset_hub::msg::QueryMsg: unknown variant")
}

func TestLocales(t *testing.T) {
	suite.Run(t, new(CollectorTestSuite))

}
