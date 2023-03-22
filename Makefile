SRC_PATH?=./cmd/theredshirts-lobby
APP_NAME?=theredshirts-lobby
DOCKER_COMPOSE_PATH?=./deployments/docker-compose-postgres.yml

init.token:
	sh ./scripts/generateKeyFile.sh

app.build:
	go mod download
	go build -o $(APP_NAME) $(SRC_PATH)

app.jt.run:
	docker compose --file $(DOCKER_COMPOSE_PATH) up --build --force-recreate -d
	go test ./test
	docker compose --file $(DOCKER_COMPOSE_PATH) down

docker.compose.run:
	docker compose --file $(DOCKER_COMPOSE_PATH) up --build