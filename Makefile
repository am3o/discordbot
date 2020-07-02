.PHONY: default test vendor

default: build

clean:
	@echo ">> Running cleaning process ..."
	rm -rf bin

test:
	@echo ">> Running tests ..."
	$(TEST_ENV_PRE) $(GO) test -cover  -race  --short $$($(GO) list ./... | grep -v /vendor/ | grep -v /integration | tr "\n" " ") $(TEST_ENV_POST)

vendor:
	@echo ">> running vendorring ..."
	go mod vendor

lint:
	@echo ">> Running linter ..."
	golangci-lint --color=always run ./...

fmt:
	@echo ">> Running code formating ..."
	go fmt $(go list ./... | grep -v /vendor/)

build: clean
	@echo ">> Running build process ... "
	go build -o bin/dircordbot .


