package utils

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"os"
	"strings"

	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client"
	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/factory"

	"github.com/lidofinance/terra-monitors/internal/app/config"

	"github.com/sirupsen/logrus"
)

const HRPAccount = "terra"

// SourceToEndpoints parses the source parameters and creates a list of endpoints based on it.
func SourceToEndpoints(source config.Source) []factory.Endpoint {
	endpoints := make([]factory.Endpoint, 0, len(source.Endpoints))
	for _, endpoint := range source.Endpoints {
		endpoints = append(endpoints, factory.Endpoint{
			Host:    endpoint,
			Schemes: source.Schemes,
		})
	}
	return endpoints
}

// BuildClient based on the endpoints, creates either a simple terra REST API client or a failover one.
func BuildClient(endpoints []factory.Endpoint, logger *logrus.Logger) *client.TerraRESTApis {
	if len(endpoints) == 1 {
		return factory.NewClient(endpoints[0], client.DefaultBasePath)
	}
	return factory.NewFailoverClient(logger, endpoints, client.DefaultBasePath)
}

func GetTerraMonitorsPath() (string, error) {
	dir, err := getCurrentDir()
	if err != nil {
		return "", err
	}

	return getTerraMonitorsPath(dir), nil
}

func ValoperToAccAddress(valoper string) (string, error) {
	_, data, err := bech32.DecodeAndConvert(valoper)
	if err != nil {
		return "", fmt.Errorf("failed to decode valoper address: %w", err)
	}

	acc, err := bech32.ConvertAndEncode(HRPAccount, data)
	if err != nil {
		return "", fmt.Errorf("failed to encode terra address: %w", err)
	}

	return acc, nil
}

func StringsSetsDifference(a, b []string) []string {
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

func getCurrentDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return dir, nil
}

func getTerraMonitorsPath(dir string) string {
	const dirName = "internal/"

	path := strings.Split(dir, dirName)

	return fmt.Sprintf("%stests/", path[0])
}
