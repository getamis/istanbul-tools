CURDIR = $(shell pwd)
#GOPATH= $(dir $(abspath $(dir $(abspath $(dir ${CURDIR})))))
GOBIN = $(CURDIR)/build/bin
GO ?= latest
VERSION ?= undefined
OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)
LDFLAGS = -s -w -X main.Version=$(VERSION)
ifeq (linux,$(OS))
	LDFLAGS+= -linkmode external -extldflags "-static"
endif

istanbul:
	@GOPATH=$(GOPATH) go build -v -o ./build/bin/istanbul ./cmd/istanbul
	@echo "Done building."
	@echo "Run \"$(GOBIN)/istanbul\" to launch istanbul."

qbft:
	@GOPATH=$(GOPATH) go build -v -o ./build/bin/qbft ./cmd/qbft
	@echo "Done building."
	@echo "Run \"$(GOBIN)/qbft\" to launch qbft."

dist: clean
	@GOPATH=$(GOPATH) go build -ldflags='$(LDFLAGS)' -o ./build/bin/istanbul ./cmd/istanbul
	@tar cfvz ./build/istanbul-tools_$(VERSION)_$(OS)_$(ARCH).tar.gz -C ./build/bin istanbul
	@echo "Distribution file created."
	@ls -lh ./build

load-testing:
	@echo "Run load testing"
	@CURDIR=$(CURDIR) go test -v github.com/jpmorganchase/istanbul-tools/tests/load/... --timeout 1h

clean:
	rm -rf build
