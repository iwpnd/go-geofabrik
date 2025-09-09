.PHONY: test
test:
	@echo "run tests"
	@go test $(go list ./... | grep -v /cmd/) -v -json | tparse -all

.PHONY: lint
lint:
	@echo "run lint"
	@golangci-lint run


