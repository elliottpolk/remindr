BIN=remindr
PKG=github.com/elliottpolk/remindr
VERSION=`cat .version`
GOOS?=linux
BUILD_DIR=./build/bin

M = $(shell printf "\033[34;1m◉\033[0m")

default: clean build ;                                              @ ## defaulting to clean and build

.PHONY: all
all: clean build

.PHONY: clean
clean: ; $(info $(M) running clean ...)                             @ ## clean up the old build dir
	@rm -vrf build

.PHONY: test
test: unit-test;													@ ## wrapper to run all testing

.PHONY: unit-test
unit-test: ; $(info $(M) running unit tests...)                     @ ## run the unit tests
	@go get -v -u
	@go test -cover ./...

.PHONY: build
build: build-dir; $(info $(M) building ...)                         @ ## build the binary
	@GOOS=$(GOOS) go build \
		-ldflags "-X main.version=$(VERSION) -X main.compiled=$(date +%s)" \
		-o ./build/bin/$(BIN) ./

.PHONY: package
package: ; $(info $(M) packaging ...)                               @ ## package up the binary for distribution to Artifactory or PCF
ifeq ($(PACKAGER),zip)
	@cd ./build/bin/ && zip $(BIN).zip $(shell ls -A ./build/bin) && rm $(BIN) && cd -
else
	@cd ./build/bin/ && tar zcvf $(BIN).tar.gz $(shell ls -A ./build/bin) && rm $(BIN) && cd -
endif

.PHONY: distro
distro: ;                                          					@ ## build and package in a distro dir for each OS
	@printf "\033[34;1m◉\033[0m cleaning up ...\n" \
		&& rm -vrf dist; mkdir dist
	@printf "\033[34;1m◉\033[0m building for Linux ...\n" \
		&& GOOS=linux $(MAKE) clean build package \
		&& mv ./build/bin/$(BIN).tar.gz dist/$(BIN)-v$(VERSION).linux.tar.gz
	@printf "\033[34;1m◉\033[0m building for macOS ...\n" \
		&& GOOS=darwin $(MAKE) clean build package \
		&& mv ./build/bin/$(BIN).tar.gz dist/$(BIN)-v$(VERSION).macos.tar.gz
	@printf "\033[34;1m◉\033[0m building for Windows ...\n" \
		&& GOOS=windows $(MAKE) clean build package \
		&& mv ./build/bin/$(BIN).tar.gz dist/$(BIN)-v$(VERSION).windows.tar.gz
	@$(MAKE) clean

.PHONY: build-dir
build-dir: ;
	@[ ! -d "${BUILD_DIR}" ] && mkdir -vp "${BUILD_DIR}" || true

PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

