DOCKER_COMPOSE_RUN ?= docker-compose

TEST_CMD ?= go test -v -p 1 -tags=integration ./...
WITH_COVER ?= -coverprofile=/bin/cover.out && go tool cover -html=/bin/cover.out -o /bin/cover.html


.PHONY: run
run: createdb
	go run cmd/main.go

.PHONY: binary
binary: 
	go build -v -o ./bin/app cmd/main.go

.PHONY: deps
deps:
	go mod vendor

.PHONY: lint
lint:
	@docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:latest golangci-lint run

.PHONY: build
build: ## Down infra
	${DOCKER_COMPOSE_RUN} build --no-cache

.PHONY: up
up: ## Up infra for tests
	${DOCKER_COMPOSE_RUN} up app mongo
	
.PHONY: down
down: ## Down infra
	${DOCKER_COMPOSE_RUN} down

.PHONY: createdb
createdb: ## Down infra
	${DOCKER_COMPOSE_RUN} up -d mongo

.PHONY: int-test
int-test: bin/ down ## Run tests
	docker-compose up -d
	go test -p 1 -tags=integration ./... -coverprofile=./cover/cover.out && go tool cover -html=./cover/cover.out -o ./cover/cover.html
	docker-compose down

.PHONY: test
test:
	go test ./... -coverprofile=./cover/cover.out && go tool cover -html=./cover/cover.out -o ./cover/cover.html
