package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/client"
	"github.com/lidofinance/terra-fcd-rest-client/columbus-5/factory"

	"github.com/lidofinance/terra-monitors/internal/app/config"

	"github.com/sirupsen/logrus"
)

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
