package client

import (
	"fmt"

	"github.com/go-openapi/runtime"
	openapiTransport "github.com/go-openapi/runtime/client"
	"github.com/lidofinance/terra-monitors/collector/config"
	"github.com/sirupsen/logrus"

	openapiClient "github.com/lidofinance/terra-monitors/openapi/client"
	openapiClientBombay "github.com/lidofinance/terra-monitors/openapi/client_bombay"
)

func New(lcd config.LCD, logger *logrus.Logger) *openapiClient.TerraLiteForTerra {
	return openapiClient.New(NewFailoverTransport(logger, lcd, openapiClient.DefaultBasePath), nil)
}

func NewBombay(lcd config.LCD, logger *logrus.Logger) *openapiClientBombay.TerraLiteForTerra {
	return openapiClientBombay.New(NewFailoverTransport(logger, lcd, openapiClient.DefaultBasePath), nil)
}

type FailoverTransport struct {
	lcd       config.LCD
	logger    *logrus.Logger
	endpoints []runtime.ClientTransport
}

func NewFailoverTransport(logger *logrus.Logger, lcd config.LCD, basepath string) runtime.ClientTransport {
	out := &FailoverTransport{
		lcd:    lcd,
		logger: logger,
	}

	for _, endpoint := range lcd.Endpoints {
		out.endpoints = append(
			out.endpoints,
			openapiTransport.New(endpoint, basepath, lcd.Schemes),
		)
	}

	return out
}

func (t *FailoverTransport) Submit(operation *runtime.ClientOperation) (resp interface{}, err error) {
	// This is a naive implementation (successively try all endpoints until one of them works).
	// Nothing fancy, but it does its job.
	for endpointID, transport := range t.endpoints {
		resp, err = transport.Submit(operation)
		if err != nil {
			t.logger.Errorf("failed to Submit to endpoint #%d (%s): %s",
				endpointID, t.lcd.Endpoints[endpointID], err)
			continue
		}

		return
	}

	return resp, fmt.Errorf("failed to Submit (all retries failed): %w", err)
}
