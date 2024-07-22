GITCOMMIT := $(shell git rev-parse HEAD)
GITDATE := $(shell git show -s --format='%ct')

LDFLAGSSTRING +=-X main.GitCommit=$(GITCOMMIT)
LDFLAGSSTRING +=-X main.GitDate=$(GITDATE)
LDFLAGS := -ldflags "$(LDFLAGSSTRING)"

lithosphere:
	env GO111MODULE=on go build -v $(LDFLAGS) ./cmd/lithosphere

clean:
	rm lithosphere

test:
	go test -v ./...

lint:
	golangci-lint run ./...

.PHONY: \
	lithosphere \
	bindings \
	bindings-scc \
	clean \
	test \
	lint
