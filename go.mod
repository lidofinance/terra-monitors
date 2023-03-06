module github.com/lidofinance/terra-monitors

go 1.16

require (
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/cosmos/cosmos-sdk v0.44.4
	github.com/gorilla/mux v1.8.0
	github.com/lidofinance/terra-fcd-rest-client v0.0.0-20220512130920-2131001551bd
	github.com/lidofinance/terra-repositories v0.0.0-20211216152128-33a198aeb9d9
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/procfs v0.7.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/vrischmann/envconfig v1.3.0
	golang.org/x/net v0.7.0 // indirect
)

// we need this replaces due to
// "Seems like current version of Go does not support replace directives in go.mod when using go get."
// https://github.com/medibloc/panacea-core/issues/198
// https://github.com/tendermint/starport/issues/155
replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/tendermint/tendermint => github.com/tendermint/tendermint v0.34.12

replace github.com/99designs/keyring => github.com/cosmos/keyring v1.1.7-0.20210622111912-ef00f8ac3d76
