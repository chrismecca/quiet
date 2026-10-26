# @(#)quiet / make / Makefile / 2026-10-25

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

WORKFLOWFILES := $(wildcard .github/workflows/*.yml)

HEADER_FILES := \
	.gitignore \
	CHANGELOG.adoc \
	CONTRIBUTING.adoc \
	Makefile \
	PORTING.adoc \
	README.adoc \
	ports/emacs/quiet-themes.el \
	$(GOFILES) \
	$(DOCFILES) \
	$(PORTFILES) \
	$(WORKFLOWFILES)

TERMINAL_APP_DIR := ports/terminal/terminal-app
TERMINAL_APP_DAY := $(TERMINAL_APP_DIR)/quiet-day.terminal
TERMINAL_APP_NIGHT := $(TERMINAL_APP_DIR)/quiet-night.terminal

# Portable generated ports, kept in the same canonical order as quiet's port
# registry. Terminal.app remains separate because its serializer requires macOS.
GENERATED_PORTS := \
	alacritty \
	base16 \
	base24 \
	chrome \
	delta \
	dircolors \
	emacs \
	firefox \
	fzf \
	ghostty \
	helix \
	iterm2 \
	jetbrains \
	kitty \
	machine \
	neovim \
	tmtheme \
	tmux \
	vim \
	vscode \
	web \
	wezterm \
	x11 \
	xcode \
	zed

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
			echo "missing quiet identification header: $$file"; \
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
generate: $(GENERATED_PORTS) report-data

# Reports ------------------------------------------------------------------

REPORTS := contrast wash cvd

# Report tables are generated Org data included by the authored report
# sources. They need only Go, so CI verifies them with the ports.
.PHONY: report-data
report-data:
	$(GO) run ./cmd/quiet report reports/data $(CONTRACT) $(PALETTES)

# PDF export needs Emacs and LuaLaTeX. SOURCE_DATE_EPOCH pins embedded dates
# to the last commit so unchanged sources export unchanged dates.
.PHONY: reports
reports: report-data
	mkdir -p reports/pdf
	cd reports/src && for report in $(REPORTS); do \
		SOURCE_DATE_EPOCH="$$(git log -1 --format=%ct)" FORCE_SOURCE_DATE=1 \
		emacs --batch -Q \
			--eval "(require 'ox-latex)" \
			--eval "(setq org-latex-pdf-process '(\"latexmk -lualatex -interaction=nonstopmode -halt-on-error %f\"))" \
			"$$report.org" \
			--funcall org-latex-export-to-pdf \
		|| exit 1; \
		mv "$$report.pdf" ../pdf/; \
	done

# Generated ports -----------------------------------------------------------

.PHONY: alacritty
alacritty:
	$(GO) run ./cmd/quiet port alacritty \
		spec/quiet-day.json \
		> ports/alacritty/quiet-day.toml
	$(GO) run ./cmd/quiet port alacritty \
		spec/quiet-night.json \
		> ports/alacritty/quiet-night.toml

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

.PHONY: chrome
chrome:
	mkdir -p ports/chrome/quiet-day ports/chrome/quiet-night
	$(GO) run ./cmd/quiet port chrome \
		spec/quiet-day.json \
		> ports/chrome/quiet-day/manifest.json
	$(GO) run ./cmd/quiet port chrome \
		spec/quiet-night.json \
		> ports/chrome/quiet-night/manifest.json

.PHONY: delta
delta:
	$(GO) run ./cmd/quiet port delta \
		spec/quiet-day.json \
		spec/quiet-night.json \
		> ports/delta/quiet.gitconfig

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

.PHONY: firefox
firefox:
	mkdir -p ports/firefox/quiet-day ports/firefox/quiet-night
	$(GO) run ./cmd/quiet port firefox \
		spec/quiet-day.json \
		> ports/firefox/quiet-day/manifest.json
	$(GO) run ./cmd/quiet port firefox \
		spec/quiet-night.json \
		> ports/firefox/quiet-night/manifest.json

.PHONY: fzf
fzf:
	$(GO) run ./cmd/quiet port fzf \
		spec/quiet-day.json \
		> ports/fzf/quiet-day.fzfrc
	$(GO) run ./cmd/quiet port fzf \
		spec/quiet-night.json \
		> ports/fzf/quiet-night.fzfrc

.PHONY: ghostty
ghostty:
	$(GO) run ./cmd/quiet port ghostty \
		spec/quiet-day.json \
		> ports/ghostty/quiet-day
	$(GO) run ./cmd/quiet port ghostty \
		spec/quiet-night.json \
		> ports/ghostty/quiet-night

.PHONY: helix
helix:
	$(GO) run ./cmd/quiet port helix \
		spec/quiet-day.json \
		> ports/helix/quiet-day.toml
	$(GO) run ./cmd/quiet port helix \
		spec/quiet-night.json \
		> ports/helix/quiet-night.toml

.PHONY: iterm2
iterm2:
	$(GO) run ./cmd/quiet port iterm2 \
		spec/quiet-day.json \
		spec/quiet-night.json \
		> ports/iterm2/quiet.itermcolors
	$(GO) run ./cmd/quiet port iterm2-variant \
		spec/quiet-day.json \
		> ports/iterm2/quiet-day.itermcolors
	$(GO) run ./cmd/quiet port iterm2-variant \
		spec/quiet-night.json \
		> ports/iterm2/quiet-night.itermcolors

.PHONY: jetbrains
jetbrains:
	mkdir -p ports/jetbrains
	$(GO) run ./cmd/quiet port jetbrains-scheme \
		spec/quiet-day.json \
		> ports/jetbrains/quiet-day.icls
	$(GO) run ./cmd/quiet port jetbrains-scheme \
		spec/quiet-night.json \
		> ports/jetbrains/quiet-night.icls
	$(GO) run ./cmd/quiet port jetbrains-theme \
		spec/quiet-day.json \
		> ports/jetbrains/quiet-day.theme.json
	$(GO) run ./cmd/quiet port jetbrains-theme \
		spec/quiet-night.json \
		> ports/jetbrains/quiet-night.theme.json
	$(GO) run ./cmd/quiet port jetbrains-plugin \
		spec/quiet-day.json \
		spec/quiet-night.json \
		> ports/jetbrains/quiet.jar

.PHONY: kitty
kitty:
	$(GO) run ./cmd/quiet port kitty \
		spec/quiet-day.json \
		> ports/kitty/quiet-day.conf
	$(GO) run ./cmd/quiet port kitty \
		spec/quiet-night.json \
		> ports/kitty/quiet-night.conf

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
		> ports/neovim/colors/quiet-day.lua
	$(GO) run ./cmd/quiet port neovim \
		spec/quiet-night.json \
		> ports/neovim/colors/quiet-night.lua

.PHONY: tmtheme
tmtheme:
	$(GO) run ./cmd/quiet port tmtheme \
		spec/quiet-day.json \
		> ports/tmtheme/quiet-day.tmTheme
	$(GO) run ./cmd/quiet port tmtheme \
		spec/quiet-night.json \
		> ports/tmtheme/quiet-night.tmTheme

.PHONY: tmux
tmux:
	$(GO) run ./cmd/quiet port tmux \
		spec/quiet-day.json \
		> ports/tmux/quiet-day.conf
	$(GO) run ./cmd/quiet port tmux \
		spec/quiet-night.json \
		> ports/tmux/quiet-night.conf

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
		> ports/vim/colors/quiet-day.vim
	$(GO) run ./cmd/quiet port vim \
		spec/quiet-night.json \
		> ports/vim/colors/quiet-night.vim

.PHONY: vscode
vscode:
	$(GO) run ./cmd/quiet port vscode \
		spec/quiet-day.json \
		> ports/vscode/themes/quiet-day-color-theme.json
	$(GO) run ./cmd/quiet port vscode \
		spec/quiet-night.json \
		> ports/vscode/themes/quiet-night-color-theme.json
	cp LICENSE ports/vscode/LICENSE

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
		> ports/wezterm/quiet-day.toml
	$(GO) run ./cmd/quiet port wezterm \
		spec/quiet-night.json \
		> ports/wezterm/quiet-night.toml

.PHONY: x11
x11:
	$(GO) run ./cmd/quiet port x11 \
		spec/quiet-day.json \
		> ports/x11/quiet-day.Xresources
	$(GO) run ./cmd/quiet port x11 \
		spec/quiet-night.json \
		> ports/x11/quiet-night.Xresources

.PHONY: xcode
xcode:
	$(GO) run ./cmd/quiet port xcode \
		spec/quiet-day.json \
		> ports/xcode/quiet-day.xcworkspacecolortheme
	$(GO) run ./cmd/quiet port xcode \
		spec/quiet-night.json \
		> ports/xcode/quiet-night.xcworkspacecolortheme

.PHONY: zed
zed:
	$(GO) run ./cmd/quiet port zed \
		spec/quiet-day.json \
		spec/quiet-night.json \
		> ports/zed/quiet.json
