GO			:= go
GO_SOURCES	:= $(shell find . -name "*.go" -type f)
GOOS		?= linux
GOARCH		?= amd64
VERSION		?= v1.0.0

.PHONY: run all fmt clean updatedep

run: all
	./bin/mirai-chess-bot-$(GOOS)-$(GOARCH)-$(VERSION)

release: all
	./scripts/release.sh $(VERSION)

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
	-rm -rvf release/*.tar.gz

updatedep:
	go mod tidy -compat=1.17
