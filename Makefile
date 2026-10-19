# @(#)quiet / make / Makefile / 2026-10-19

GO := go
BIN := bin/quiet

CONTRACT := spec/quiet-core.json
PALETTES := \
	spec/quiet-day.json \
	spec/quiet-night.json

GOFILES := $(shell find cmd internal -type f -name '*.go' -print | sort)
DOCFILES := $(shell find docs spec reports/src -type f \
	\( -name '*.adoc' -o -name '*.org' \) -print 2>/dev/null | sort)
PORTFILES := $(shell find ports -type f \
	\( -name '*.adoc' -o -name '*.org' -o -name '*.swift' -o -name '*.sh' \) \
	-print 2>/dev/null | sort)

HEADER_FILES := \
	.gitignore \
	CONTRIBUTING.adoc \
	Makefile \
	README.adoc \
	$(GOFILES) \
	$(DOCFILES) \
	$(PORTFILES)

TERMINAL_APP_DIR := ports/terminal/terminal-app
TERMINAL_APP_DAY := $(TERMINAL_APP_DIR)/quiet-day.terminal
TERMINAL_APP_NIGHT := $(TERMINAL_APP_DIR)/quiet-night.terminal

# Portable generated ports, kept in the same canonical order as quiet's port
# registry. Terminal.app remains separate because its serializer requires macOS.
GENERATED_PORTS := \
	base16 \
	base24 \
	dircolors \
	emacs \
	ghostty \
	iterm2 \
	kitty \
	machine \
	neovim \
	vim \
	vscode \
	web \
	wezterm \
	xcode

.DEFAULT_GOAL := check

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
	rm -rf bin coverage.out tmp

.PHONY: generate
generate: $(GENERATED_PORTS)

# Generated ports -----------------------------------------------------------

.PHONY: base16
base16:
	$(GO) run ./cmd/quiet port base16 \
		spec/quiet-day.json \
		> ports/base16/quiet-day.yaml
	$(GO) run ./cmd/quiet port base16 \
		spec/quiet-night.json \
		> ports/base16/quiet-night.yaml

.PHONY: base24
base24:
	$(GO) run ./cmd/quiet port base24 \
		spec/quiet-day.json \
		> ports/base24/quiet-day.yaml
	$(GO) run ./cmd/quiet port base24 \
		spec/quiet-night.json \
		> ports/base24/quiet-night.yaml

.PHONY: dircolors
dircolors:
	$(GO) run ./cmd/quiet port dircolors \
		> ports/dircolors/quiet.dircolors

.PHONY: emacs
emacs:
	$(GO) run ./cmd/quiet port emacs \
		spec/quiet-day.json \
		> ports/emacs/quiet-day-theme.el
	$(GO) run ./cmd/quiet port emacs \
		spec/quiet-night.json \
		> ports/emacs/quiet-night-theme.el

.PHONY: ghostty
ghostty:
	$(GO) run ./cmd/quiet port ghostty \
		spec/quiet-day.json \
		> 'ports/ghostty/Quiet Day'
	$(GO) run ./cmd/quiet port ghostty \
		spec/quiet-night.json \
		> 'ports/ghostty/Quiet Night'

.PHONY: iterm2
iterm2:
	$(GO) run ./cmd/quiet port iterm2 \
		spec/quiet-day.json \
		spec/quiet-night.json \
		> 'ports/iterm2/Quiet.itermcolors'

.PHONY: kitty
kitty:
	$(GO) run ./cmd/quiet port kitty \
		spec/quiet-day.json \
		> 'ports/kitty/Quiet Day.conf'
	$(GO) run ./cmd/quiet port kitty \
		spec/quiet-night.json \
		> 'ports/kitty/Quiet Night.conf'

.PHONY: machine
machine:
	$(GO) run ./cmd/quiet port machine \
		spec/quiet-day.json \
		spec/quiet-night.json \
		> ports/machine/quiet.json

.PHONY: neovim
neovim:
	$(GO) run ./cmd/quiet port neovim \
		spec/quiet-day.json \
		> ports/neovim/quiet-day.lua
	$(GO) run ./cmd/quiet port neovim \
		spec/quiet-night.json \
		> ports/neovim/quiet-night.lua

.PHONY: terminal-app-smoke
terminal-app-smoke:
	@if [ "$$(uname -s)" != "Darwin" ]; then \
		echo "Terminal.app smoke test requires macOS"; \
		exit 2; \
	fi
	mkdir -p tmp
	swift $(TERMINAL_APP_DIR)/generate.swift \
		$(TERMINAL_APP_DIR)/testdata/sentinel.json \
		tmp/quiet-terminal-sentinel.terminal
	plutil -lint tmp/quiet-terminal-sentinel.terminal

.PHONY: terminal-app-profiles
terminal-app-profiles: build
	@if [ "$$(uname -s)" != "Darwin" ]; then \
		echo "Terminal.app profile generation requires macOS"; \
		exit 2; \
	fi
	mkdir -p tmp
	$(BIN) port terminal \
		spec/quiet-day.json \
		> tmp/quiet-day-terminal.json
	$(BIN) port terminal \
		spec/quiet-night.json \
		> tmp/quiet-night-terminal.json
	swift $(TERMINAL_APP_DIR)/generate.swift \
		tmp/quiet-day-terminal.json \
		$(TERMINAL_APP_DAY)
	swift $(TERMINAL_APP_DIR)/generate.swift \
		tmp/quiet-night-terminal.json \
		$(TERMINAL_APP_NIGHT)
	plutil -lint \
		$(TERMINAL_APP_DAY) \
		$(TERMINAL_APP_NIGHT)

.PHONY: vim
vim:
	$(GO) run ./cmd/quiet port vim \
		spec/quiet-day.json \
		> ports/vim/quiet-day.vim
	$(GO) run ./cmd/quiet port vim \
		spec/quiet-night.json \
		> ports/vim/quiet-night.vim

.PHONY: vscode
vscode:
	$(GO) run ./cmd/quiet port vscode \
		spec/quiet-day.json \
		> ports/vscode/themes/quiet-day-color-theme.json
	$(GO) run ./cmd/quiet port vscode \
		spec/quiet-night.json \
		> ports/vscode/themes/quiet-night-color-theme.json

.PHONY: web
web:
	$(GO) run ./cmd/quiet port web \
		spec/quiet-day.json \
		spec/quiet-night.json \
		> ports/web/quiet.css

.PHONY: wezterm
wezterm:
	$(GO) run ./cmd/quiet port wezterm \
		spec/quiet-day.json \
		> 'ports/wezterm/Quiet Day.toml'
	$(GO) run ./cmd/quiet port wezterm \
		spec/quiet-night.json \
		> 'ports/wezterm/Quiet Night.toml'

.PHONY: xcode
xcode:
	$(GO) run ./cmd/quiet port xcode \
		spec/quiet-day.json \
		> 'ports/xcode/Quiet Day.xcworkspacecolortheme'
	$(GO) run ./cmd/quiet port xcode \
		spec/quiet-night.json \
		> 'ports/xcode/Quiet Night.xcworkspacecolortheme'
