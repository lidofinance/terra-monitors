gen_client:
	mkdir -p ./openapi/ && swagger generate client -f swagger.yaml -t ./openapi/

dev_server: gen_client
	go run main.go

start:
	docker-compose --env-file ./docker/env/.lido_terra.env up -d --build

start_testnet:
	docker-compose --env-file ./docker/env/.lido_terra.testnet.env up --build -d

start_mainnet:
	docker-compose --env-file ./docker/env/.lido_terra.prod.env -p terra_monitoring_mainnet up --build -d

start-no-build:
	docker-compose up -d

stop:
	docker-compose down --remove-orphans

stop_testnet:
	docker-compose down --remove-orphans

stop_mainnet:
	docker-compose -p terra_monitoring_mainnet down --remove-orphans

test:
	go test ./...
