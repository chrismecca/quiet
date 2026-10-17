# @(#)quiet / make / Makefile / 2026-10-16

GO := go
BIN := bin/quiet

CONTRACT := spec/quiet-core.json
PALETTES := \
	spec/quiet-day.json \
	spec/quiet-night.json

GOFILES := $(shell find cmd internal -type f -name '*.go' -print | sort)
DOCFILES := $(shell find docs spec reports/src -type f \
	\( -name '*.adoc' -o -name '*.org' \) -print 2>/dev/null | sort)

HEADER_FILES := \
	.gitignore \
	CONTRIBUTING.adoc \
	Makefile \
	README.adoc \
	$(GOFILES) \
	$(DOCFILES) \
	$(PORTFILES)

.DEFAULT_GOAL := check

PORTFILES := $(shell find ports -type f \
	\( -name '*.adoc' -o -name '*.org' -o -name '*.swift' -o -name '*.sh' \) \
	-print 2>/dev/null | sort)

.PHONY: all
all: check build

.PHONY: fmt
fmt:
	gofmt -w $(GOFILES)

.PHONY: fmt-check
fmt-check:
	@files="$$(gofmt -l $(GOFILES))"; \
	if [ -n "$$files" ]; then \
		echo "Go files need formatting:"; \
		echo "$$files"; \
		exit 1; \
	fi

.PHONY: headers
headers:
	@status=0; \
	for file in $(HEADER_FILES); do \
		if ! head -n 3 "$$file" | grep -Fq '@(#)quiet /'; then \
			echo "missing Quiet identification header: $$file"; \
			status=1; \
		fi; \
	done; \
	exit $$status

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: test
test:
	$(GO) test ./...

.PHONY: core-check
core-check:
	$(GO) run ./cmd/quiet check $(CONTRACT) $(PALETTES)

.PHONY: check
check: headers fmt-check vet test core-check

.PHONY: build
build:
	mkdir -p bin
	$(GO) build -trimpath -o $(BIN) ./cmd/quiet

.PHONY: coverage
coverage:
	$(GO) test -coverprofile=coverage.out ./...

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: clean
clean:
	rm -rf bin coverage.out

.PHONY: terminal-app-smoke
terminal-app-smoke:
	@if [ "$$(uname -s)" != "Darwin" ]; then \
		echo "Terminal.app smoke test requires macOS"; \
		exit 2; \
	fi
	mkdir -p tmp
	swift ports/terminal/terminal-app/generate.swift \
		ports/terminal/terminal-app/testdata/sentinel.json \
		tmp/quiet-terminal-sentinel.terminal
	plutil -lint tmp/quiet-terminal-sentinel.terminal

