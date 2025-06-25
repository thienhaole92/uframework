.PHONY:pre-lint
pre-lint:
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY:lint
lint: pre-lint
	go mod tidy
	gofumpt -l -w .
	go vet ./...
	golangci-lint run

.PHONY: test
test:
	go test ./... -v -coverprofile=coverage.out

.PHONY:pre-benchmark
pre-benchmark:
	go install golang.org/x/perf/cmd/benchstat@latest

.PHONY:benchmark
benchmark: pre-benchmark
	go test ./... -bench=. -count=6 > benchmark.out

.PHONY:benchstat
benchstat: benchmark
	benchstat benchmark.out
