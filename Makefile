.PHONY: build-linux clean

# ====================================================================================
# Build Commands
# ====================================================================================

## build-linux: Builds the distributable .deb package for Linux.
build-linux:
	@echo "--> Running Linux build script..."
	@chmod +x ./buildScript/buildDeb.sh
	@./buildScript/buildDeb.sh

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

## help: Shows this help message.
help:
	@echo "Usage: make <command>"
	@echo ""
	@echo "Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help 