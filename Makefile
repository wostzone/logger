DIST_FOLDER=./dist
.DEFAULT_GOAL := help

.PHONY: help

all: logger ## Build package with binary distribution and config

install:  ## Install the plugin into ~/bin/wost/bin and config
	cp dist/bin/* ~/bin/wost/bin/
	cp dist/arm/* ~/bin/wost/arm/
	cp -n dist/config/* ~/bin/wost/config/

logger:
	GOOS=linux GOARCH=amd64 go build -o $(DIST_FOLDER)/bin/$@ ./main.go
	GOOS=linux GOARCH=arm go build -o $(DIST_FOLDER)/arm/$@ ./main.go
	@echo "> SUCCESS. Plugin '$@' can be found at $(DIST_FOLDER)/bin/$@ and $(DIST_FOLDER)/arm/$@"


clean: ## Clean distribution files
	$(GOCLEAN)
	rm -f $(DIST_FOLDER)/certs/*
	rm -f $(DIST_FOLDER)/logs/*
	rm -f $(DIST_FOLDER)/bin/*
	rm -f $(DIST_FOLDER)/arm/*
	rm -f ./test/certs/*
	rm -f ./test/logs/*
	rm -f ./test/bin/*
	rm -f ./test/arm/*


help: ## Show this help
		@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
