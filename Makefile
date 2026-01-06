APP      := scode
PKG      := ./...
BIN_DIR  := bin
VERSION  := dev
GOFLAGS  := -trimpath

.PHONY: build
build:
	go build $(GOFLAGS) -o $(BIN_DIR)/$(APP)

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)

.PHONY: install
install:
	go install $(GOFLAGS)

.PHONY: release
release:
	GOOS=linux   GOARCH=amd64 go build $(GOFLAGS) -o $(BIN_DIR)/$(APP)-linux-amd64
	GOOS=darwin  GOARCH=amd64 go build $(GOFLAGS) -o $(BIN_DIR)/$(APP)-darwin-amd64
	GOOS=darwin  GOARCH=arm64 go build $(GOFLAGS) -o $(BIN_DIR)/$(APP)-darwin-arm64
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -o $(BIN_DIR)/$(APP)-windows-amd64.exe
