.PHONY: default test vendor lint fmt build

default: clean lint test build

clean:
	@echo ">> Running cleaning process ..."
	rm -rf bin

test:
	@echo ">> Running tests ..."
	go test -cover  -race  --short $$(go list ./... | grep -v /vendor/ | grep -v /integration | tr "\n" " ")

vendor:
	@echo ">> running vendoring ..."
	go mod vendor

vet: ## Run go vet against code
	go vet ./pkg/... .

lint:
	@echo ">> Running linter ..."
	golangci-lint --color=always run ./...

fmt:
	@echo ">> Running code formating ..."
	go fmt $$(go list ./... | grep -v /vendor/)

build: clean
	@echo ">> Running build process ... "
	go build -o bin/dircordbot .


