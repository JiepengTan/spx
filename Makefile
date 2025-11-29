# ============================================
# Config
# ============================================
.DEFAULT_GOAL := help

CURRENT_PATH := $(shell pwd)

# Automatically collect demos from directories
DEMOS := $(wildcard tutorial/*)
DEMO_INDEX ?= 3

PORT    ?= 8106
MOVIE   ?= false

# Command to install spx
INSTALL_CMD = cd ./cmd/gox && ./install.sh && cd $(CURRENT_PATH)


# ============================================
# Help
# ============================================
help: ## Show available commands
	echo "Detected demos: $(DEMOS)"
	@echo "Make Commands:"
	@echo "================================"
	@grep -E '^[a-zA-Z0-9._-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  make %-25s %s\n", $$1, $$2}'
	@echo ""
	@echo "Demo targets via index:"
	@i=1; \
	for demo in $(DEMOS); do \
		echo "  make run DEMO_INDEX=$$i          # Run $$demo"; \
		echo "  make run-web DEMO_INDEX=$$i      # Run web $$demo"; \
		echo "  make run-editor DEMO_INDEX=$$i   # Run editor $$demo"; \
		i=$$((i+1)); \
	done

# ============================================
# Demo Commands
# ============================================
list-demos: ## List all demos with index
	@i=1; \
	for demo in $(DEMOS); do \
		echo "$$i: $$demo"; \
		i=$$((i+1)); \
	done

# ============================================
# Setup Commands
# ============================================
setup: ## Initialize the user environment
	chmod +x ./pkg/gdspx/tools/*.sh && \
	echo "===> Step 1/4: Install spx" && \
	make install && \
	echo "===> Step 2/4: Download engine" && \
	make download && \
	echo "===> Step 3/4: Export runtime package" && \
	make export-pack && \
	echo "===> Step 4/4: Prepare web template" && \
	./pkg/gdspx/tools/make_util.sh extrawebtemplate && \
	echo "===> setup done"


setup-dev: ## Initialize development environment (full)
	chmod +x ./pkg/gdspx/tools/*.sh && \
	echo "===> Step 1/6: Install spx" && \
	make install && \
	echo "===> Step 2/6: Download engine" && \
	make download && \
	echo "===> Step 3/6: Build wasm" && \
	make build-wasm && \
	echo "===> Step 4/6: Build editor engine" && \
	make build-editor && \
	echo "===> Step 5/6: Build desktop engine" && \
	make build-desktop && \
	echo "===> Step 6/6: Build web engine" && \
	make build-web && \
	echo "===> setup-dev done, use 'make run DEMO_INDEX=N' to run demo"

setup-web: ## Download and install web engine from godot releases. Usage: make setup-web MODE=normal (MODE: normal|worker|minigame|miniprogram)
ifndef MODE
	$(error MODE is not set! Usage: make setup-web MODE=normal or MODE=worker or MODE=minigame or MODE=miniprogram)
endif
	@if [ "$(MODE)" != "normal" ] && [ "$(MODE)" != "worker" ] && [ "$(MODE)" != "minigame" ] && [ "$(MODE)" != "miniprogram" ]; then \
		echo "Error: Invalid MODE '$(MODE)'. Supported modes: normal, worker, minigame, miniprogram"; \
		exit 1; \
	fi
	echo "===> Setting up web $(MODE) engine..."
	make build-wasm && \
	./pkg/gdspx/tools/build_engine.sh -g -p web -m $(MODE) && \
	./pkg/gdspx/tools/make_util.sh extrawebtemplate $(MODE) && \
	echo "===> Web $(MODE) engine setup complete"


# ============================================
# Install & Download
# ============================================
install: ## Install spx command
	$(INSTALL_CMD)

download: ## Download engines
	make install && ./pkg/gdspx/tools/build_engine.sh -e -d

download-engine: ## Download engine templates for specific platform. Usage: make download-engine PLATFORM=android|ios|web [MODE=normal|worker|minigame|miniprogram]
ifndef PLATFORM
	$(error PLATFORM is not set! Usage: make download-engine PLATFORM=android or PLATFORM=ios or PLATFORM=web [MODE=mode])
endif
	@echo "Downloading engine templates for platform: $(PLATFORM)"
	@if [ "$(PLATFORM)" = "web" ]; then \
		if [ -n "$(MODE)" ]; then \
			./pkg/gdspx/tools/build_engine.sh -p $(PLATFORM) -g -m $(MODE); \
		else \
			./pkg/gdspx/tools/build_engine.sh -p $(PLATFORM) -g; \
		fi \
	else \
		./pkg/gdspx/tools/build_engine.sh -p $(PLATFORM) -g; \
	fi 


# ============================================
# Build Commands
# ============================================
build-editor: ## Build editor mode engine
	./pkg/gdspx/tools/build_engine.sh -e

build-desktop: ## Build desktop engine
	./pkg/gdspx/tools/build_engine.sh && \
	./pkg/gdspx/tools/make_util.sh exportpack 

build-web: ## Build web engine template
	./pkg/gdspx/tools/build_engine.sh -p web && \
	./pkg/gdspx/tools/make_util.sh extrawebtemplate normal

build-web-worker: ## Build web worker engine template
	make install && make build-wasm && \
	./pkg/gdspx/tools/build_engine.sh -p web -m worker && \
	./pkg/gdspx/tools/make_util.sh extrawebtemplate worker

build-minigame: ## Build minigame template
	./pkg/gdspx/tools/build_engine.sh -p web -m minigame && \
	./pkg/gdspx/tools/make_util.sh extrawebtemplate minigame

build-miniprogram: ## Build miniprogram template
	./pkg/gdspx/tools/build_engine.sh -p web -m miniprogram && \
	./pkg/gdspx/tools/make_util.sh extrawebtemplate miniprogram

build-wasm: ## Build wasm
	cd ./cmd/gox/ && ./install.sh --web && cd $(CURRENT_PATH)

build-wasm-opt: ## Build wasm with optimization
	cd ./cmd/gox/ && ./install.sh --web --opt && cd $(CURRENT_PATH)
	./pkg/gdspx/tools/make_util.sh compresswasm

build-android: ## Build android engine
	make install &&./pkg/gdspx/tools/build_engine.sh -p android

build-ios: ## Build ios engine
	make install &&./pkg/gdspx/tools/build_engine.sh -p ios 

# ============================================
# Run Commands (by index)
# ============================================
define GET_DEMO
$(word $(DEMO_INDEX),$(DEMOS))
endef

editor: ## Open demo in editor: make editor DEMO_INDEX=N
ifndef DEMO_INDEX
	$(error DEMO_INDEX is not set! Usage: make editor DEMO_INDEX=N)
endif
	@DEMO=$(GET_DEMO); \
	echo "Opening editor for demo #$(DEMO_INDEX): $$DEMO"; \
	cd $$DEMO && spx editor -movie=$(MOVIE)

run: ## Run demo on PC: make run DEMO_INDEX=N
ifndef DEMO_INDEX
	$(error DEMO_INDEX is not set! Usage: make run DEMO_INDEX=N)
endif
	@DEMO=$(GET_DEMO); \
	echo "Running demo #$(DEMO_INDEX): $$DEMO"; \
	cd $$DEMO && spx run -movie=$(MOVIE)

run-editor: ## Run demo in editor mode: make run-editor DEMO_INDEX=N
ifndef DEMO_INDEX
	$(error DEMO_INDEX is not set! Usage: make run-editor DEMO_INDEX=N)
endif
	@DEMO=$(GET_DEMO); \
	echo "Running editor demo #$(DEMO_INDEX): $$DEMO"; \
	cd $$DEMO && spx rune -movie=$(MOVIE)

run-web: ## Run demo on web: make run-web DEMO_INDEX=N
ifndef DEMO_INDEX
	$(error DEMO_INDEX is not set! Usage: make run-web DEMO_INDEX=N)
endif
	@DEMO=$(GET_DEMO); \
	echo "Running web demo #$(DEMO_INDEX): $$DEMO"; \
	make stop &&  \
	cd $$DEMO && spx clear && spx runweb -serveraddr=":$(PORT)"

run-web-worker: ## Run demo on web: make run-web DEMO_INDEX=N
ifndef DEMO_INDEX
	$(error DEMO_INDEX is not set! Usage: make run-web DEMO_INDEX=N)
endif
	@DEMO=$(GET_DEMO); \
	echo "Running web demo #$(DEMO_INDEX): $$DEMO"; \
	make stop &&  \
	cd $$DEMO && spx clear && spx runwebworker -serveraddr=":$(PORT)"
# ============================================
# Utility Commands
# ============================================
format: ## Format Go code
	go fmt ./...

generate: ## Generate code
	cd ./pkg/gdspx/cmd/codegen && go run . && cd $(CURRENT_PATH) && make format

export-pack: ## Export runtime pck file
	./pkg/gdspx/tools/make_util.sh exportpack && cd $(CURRENT_PATH)

export-web: ## Export web engine. Usage: make export-web MODE=normal (MODE: normal|worker|minigame|miniprogram)
	@MODE=$(or $(MODE),normal)
	@if [ "$(MODE)" != "normal" ] && [ "$(MODE)" != "worker" ] && [ "$(MODE)" != "minigame" ] && [ "$(MODE)" != "miniprogram" ]; then \
		echo "Error: Invalid MODE '$(MODE)'. Supported modes: normal, worker, minigame, miniprogram"; \
		exit 1; \
	fi
	cd ./cmd/gox && ./install.sh --web --opt && cd $(CURRENT_PATH) && \
	./pkg/gdspx/tools/make_util.sh exportweb $(MODE) && cd $(CURRENT_PATH)

export-web-worker: ## Export web worker engine package (deprecated, use make export-web MODE=worker)
	@echo "===> Exporting web worker package..."
	@cd ./cmd/gox && ./install.sh --web --opt && cd $(CURRENT_PATH) && \
	./pkg/gdspx/tools/make_util.sh exportweb worker && cd $(CURRENT_PATH)
	@echo "===> Web worker export complete - spx_web_worker.zip created"

stop: ## Stop running processes
	@echo "Stopping running processes..."
	@if [ "$$OS" = "Windows_NT" ]; then \
		taskkill /F /FI "IMAGENAME eq python.exe" 2>NUL || true; \
		taskkill /F /FI "IMAGENAME eq python3.exe" 2>NUL || true; \
	else \
		PIDS=$$(pgrep -f gdspx_web_server.py || true); \
		if [ -n "$$PIDS" ]; then kill -9 $$PIDS; fi \
	fi
	@echo "Processes stopped."

