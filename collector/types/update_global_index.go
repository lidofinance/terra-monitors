package types

type UpdateGlobalIndexBotState struct {
	SuccessfulTxSinceLastCheck float64
	GasWantedSinceLastCheck    float64
	GasUsedSinceLastCheck      float64
	UUSDFeeSinceLastCheck      float64
}
