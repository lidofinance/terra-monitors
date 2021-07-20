gen_client:
	swagger generate client -f swagger.yaml

start:
	docker-compose up -d --build

stop:
	docker-compose down --remove-orphans