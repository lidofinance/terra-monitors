gen_client:
	swagger generate client -f swagger.yaml

dev_server: gen_client
	go run main.go

start:
	docker-compose up -d --build

stop:
	docker-compose down --remove-orphans


test:
	go test ./...
