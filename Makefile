GOBIN = build/bin
GO ?= latest

istanbul:
	go build -v -o ./build/bin/istanbul ./cmd/istanbul
	@echo "Done building."
	@echo "Run \"$(GOBIN)/istanbul\" to launch istanbul."

clean:
	rm -rf build/bin/