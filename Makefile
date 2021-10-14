gen_client: gen_columbus_4 gen_columbus_5

gen_columbus_4:
	mkdir -p ./openapi/ && swagger generate client -f swagger.yaml -t ./openapi/


gen_columbus_5:
	mkdir -p ./openapi/ && swagger generate client -f swagger.bombay.yaml -t ./openapi/ -c client_bombay

dev_server: gen_client
	go run ./cmd/terra-monitors/main.go

start:
	docker-compose --env-file ./docker/env/.lido_terra.env up -d --build

start_testnet:
	docker-compose --env-file ./docker/env/.lido_terra.testnet.env up --build -d

start_mainnet:
	docker-compose --env-file ./docker/env/.lido_terra.prod.env -p terra-monitoring-mainnet up --build -d

start-no-build:
	docker-compose up -d

stop:
	docker-compose down --remove-orphans

stop_testnet:
	docker-compose down --remove-orphans

stop_mainnet:
	docker-compose -p terra-monitoring-mainnet down --remove-orphans

start_bombay:
	docker-compose --env-file ./docker/env/.lido_terra.bombay.env -p terra_monitors_bombay up --build -d

stop_bombay:
	docker-compose -p terra_monitors_bombay down --remove-orphans

test:
	go test ./...
