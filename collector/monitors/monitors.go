package monitors

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lidofinance/terra-monitors/collector/types"
	"github.com/lidofinance/terra-monitors/openapi/client"
	"github.com/lidofinance/terra-monitors/openapi/client/wasm"
	"github.com/sirupsen/logrus"
)

type Monitor interface {
	Name() string
	InitMetrics()
	// Handler fetches the data to inner storage
	Handler(ctx context.Context) error
	// GetMetrics - provides metrics fetched by Handler method
	GetMetrics() map[Metric]float64
}


type StoreQueryMonitor interface {
	Name() string
	GetApiClient() *client.TerraLiteForTerra
	GetLogger() *logrus.Logger
	GetContract() string
}

type Metric string


type BaseMonitor struct {
	ContractAddress string
	apiClient       *client.TerraLiteForTerra
	logger          *logrus.Logger
}

func (h BaseMonitor) GetApiClient() *client.TerraLiteForTerra {
	return h.apiClient
}

func (h BaseMonitor) GetLogger() *logrus.Logger {
	return h.logger
}

func (h BaseMonitor) GetContract() string {
	return h.ContractAddress
}




func makeStoreQuery(responsePtr, request interface{}, ctx context.Context, m StoreQueryMonitor) error {
	reqRaw, err := json.Marshal(&request)
	if err != nil {
		return fmt.Errorf("failed to marshal %s request: %w", m.Name(), err)
	}

	p := wasm.GetWasmContractsContractAddressStoreParams{}
	p.SetContext(ctx)
	p.SetContractAddress(m.GetContract())
	p.SetQueryMsg(string(reqRaw))

	resp, err := m.GetApiClient().Wasm.GetWasmContractsContractAddressStore(&p)
	if err != nil {
		return fmt.Errorf("failed to process %s request: %w", m.Name(), err)
	}

	err = types.CastMapToStruct(resp.Payload.Result, responsePtr)
	if err != nil {
		return fmt.Errorf("failed to parse %s body interface: %w", m.Name(), err)
	}

	m.GetLogger().Infoln("updated", m.Name())
	return nil
}
