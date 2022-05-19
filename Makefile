export GO111MODULE=on
BIN_NAME := .$(or $(PROJECT_NAME), 'app')
GOLINT := ./bin/golangci-lint

dep: # Download required dependencies
	go mod tidy
	go mod download
	go mod vendor

build: dep
	CGO_ENABLED=1 go build -mod=vendor -o ./bin/${BIN_NAME} -a ./cmd/meeting-app

run: dep
	go run ./cmd/app

clean: ## Remove previous build
	rm -f bin/$(BIN_NAME)

check-swagger:
	@which swagger || (GO111MODULE=off go get -u github.com/go-swagger/go-swagger/cmd/swagger)

swagger: check-swagger
#	swagger generate spec -o src/server/http/static/v1/swagger.yaml  -w ./ --scan-models --exclude-tag=external
#	swagger generate spec -o src/server/http/static/v1/swagger.json  -w ./ --scan-models --exclude-tag=external

check-lint:
	@which $(GOLINT) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.46.2

lint: dep check-lint
	$(GOLINT) run -c .golangci.yml --timeout 5m

lint-fix: dep check-lint
	$(GOLINT) run -c .golangci.yml --timeout 5m --fix

dc-run-infrastructure:
	@docker-compose -f dev/docker-compose.yml up -d mongodb redisdb

dc-run-dev:
	@docker-compose -f dev/docker-compose.yml up develop

dc-run-app:
	@docker-compose -f dev/docker-compose.yml up app

dc-clean: ## clean up dockerized infrastructure
	@docker-compose -f dev/docker-compose.yml stop ; docker-compose -f dev/docker-compose.yml rm -f;