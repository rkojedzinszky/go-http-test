
ARCHS = amd64 arm64 arm

BIN := $(patsubst %,bin/go-http-test.%,$(ARCHS))

all: $(BIN)

bin/go-http-test.%: main.go go.mod
	CGO_ENABLED=0 GOARCH=$(patsubst bin/go-http-test.%,%,$@) go build -ldflags -s -o $@ .

