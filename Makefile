GO			:= go
GO_SOURCES	:= $(shell find . -name "*.go" -type f)
GOOS		?= linux
GOARCH		?= amd64
VERSION		?= v0.0.1

.PHONY: run all fmt clean updatedep

run: all
	./bin/mirai-chess-bot-linux-amd64-v0.0.1

release: all
	./scripts/release.sh

all: bin/mirai-chess-bot-$(GOOS)-$(GOARCH)-$(VERSION) inkscape device

inkscape:
	./scripts/download_inkscape.sh

device: bin/device
	./bin/device

bin/device: $(GO_SOURCES)
	GOOS=$(GOOS) GOARCH=$(GOARCH) \
	$(GO) build -o bin/device \
	cmd/device/device.go

bin/mirai-chess-bot-$(GOOS)-$(GOARCH)-$(VERSION): $(GO_SOURCES)
	GOOS=$(GOOS) GOARCH=$(GOARCH) \
	$(GO) build -o bin/mirai-chess-bot-$(GOOS)-$(GOARCH)-$(VERSION) \
	cmd/bot/bot.go

fmt:
	gofmt -l -w $(GO_SOURCES)

clean:
	-rm -rvf bin/mirai-chess-bot-*
	-rm -rvf bin/device

updatedep:
	go mod tidy -compat=1.17
