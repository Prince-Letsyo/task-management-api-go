APP_NAME = app
BUILD_DIR = $(PWD)/build
CONFIG_FILE = .$(APP_NAME).config.dev.yaml


clean:
	rm -rf $(BUILD_DIR)

critic:
	gocritic check -enableAll main

security:
	gosec ./...

lint:
	golangci-lint run ./...

test: clean critic security lint
	go test -v -timeout 60s -coverprofile=cover.out -cover -p 1 ./...
	go tool cover -html=cover.out -o coverage.html

build:test
	CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/main.go

migrate:
	go run ./cmd/main.go --migrate

run: build
	$(BUILD_DIR)/$(APP_NAME)

run.go:
	go run ./cmd/main.go -c $(CONFIG_FILE)

run.air:
	air

swag:
	swag fmt -d ./internal && swag init -d ./cmd/$(APP_NAME),./internal/$(APP_NAME),./internal/controller/http -pd fiber
