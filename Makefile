GO			:= go
GO_SOURCES	:= $(shell find . -name "*.go" -type f)
GOOS		?= linux
GOARCH		?= amd64
VERSION		?= v0.0.1

.PHONY: all clean updatedep

all: bin/mirai-chess-bot-$(GOOS)-$(GOARCH)-$(VERSION)

bin/mirai-chess-bot-$(GOOS)-$(GOARCH)-$(VERSION): $(GO_SOURCES)
	GOOS=$(GOOS) GOARCH=$(GOARCH) \
	$(GO) build -o bin/mirai-chess-bot-$(GOOS)-$(GOARCH)-$(VERSION) \
	cmd/bot.go

fmt:
	gofmt -l -w $(GO_SOURCES)

clean:
	-rm -rvf bin/mirai-chess-bot-*

updatedep:
	go mod tidy -compat=1.17
