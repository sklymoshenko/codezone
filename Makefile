.PHONY: build-linux clean postgres-test

# ====================================================================================
# Build Commands
# ====================================================================================

## build-linux: Builds the distributable .deb package for Linux.
build-lin:
	@echo "--> Running Linux build script..."
	@chmod +x ./buildScript/buildDeb.sh
	@./buildScript/buildDeb.sh

## build-windows: Builds the distributable .exe package for Windows.
build-win:
	@echo "--> Running Windows build script..."
	@powershell -ExecutionPolicy Bypass -File ./buildScript/buildWin.ps1

## build-mac: Builds the distributable .app package for macOS.
build-mac:
	@echo "--> Running macOS build script..."
	@chmod +x ./buildScript/buildMac.sh
	@./buildScript/buildMac.sh

# ====================================================================================
# Utility Commands
# ====================================================================================

## clean: Removes all build artifacts and temporary files.
clean:
	@echo "--> Cleaning up build artifacts..."
	@rm -f *.deb
	@rm -rf build
	@rm -rf packaging
	@echo "Cleanup complete."

# ====================================================================================
# Test Commands
# ====================================================================================

## postgres-test: Runs PostgreSQL integration tests (requires Docker).
postgres-test:
	@./scripts/test-integration.sh all

## help: Shows this help message.
help:
	@echo "Usage: make <command>"
	@echo ""
	@echo "Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help 