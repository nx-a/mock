APP := mock
DIST := dist
GO ?= go

.PHONY: all build linux windows release test vet check clean

all: build

build:
	@mkdir -p $(DIST)
	$(GO) build -trimpath -o $(DIST)/$(APP) ./cmd/$(APP)

linux:
	@mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -trimpath -o $(DIST)/$(APP)-linux-amd64 ./cmd/$(APP)

windows:
	@mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -trimpath -o $(DIST)/$(APP)-windows-amd64.exe ./cmd/$(APP)

release: clean linux windows

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

check: vet test

clean:
	rm -rf $(DIST)
