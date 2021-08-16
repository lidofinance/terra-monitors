gen_client:
	mkdir -p ./openapi/ && swagger generate client -f swagger.yaml -t ./openapi/

dev_server: gen_client
	go run main.go

start:
	docker-compose up -d --build

start-no-build:
	docker-compose up -d

stop:
	docker-compose down --remove-orphans


test:
	go test ./...
