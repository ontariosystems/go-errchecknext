NAME = go-errchecknext
.DEFAULT_GOAL:=all

.PHONY: all
all: lint test

export GOPROXY=proxy.golang.org,direct

.PHONY: test
test:
	go run github.com/onsi/ginkgo/v2/ginkgo run \
		-r \
		--cover \
		--randomize-all \
		--coverpkg ./...

lint: errchecknext golangci-lint

.PHONY: golangci-lint
golangci-lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run

.PHONY: errchecknext
errchecknext:
	go run . ./...
