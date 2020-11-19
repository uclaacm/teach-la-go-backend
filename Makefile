BIN_OUT=bin
BIN_NAME=tlabe
PLATFORMS=windows/386 \
	windows/amd64 \
	darwin/amd64 \
	linux/386 \
	linux/amd64 \
	linux/arm

.PHONY: default
default: clean host

.PHONY: all
all: host $(PLATFORMS)

.PHONY: host
host:
	go build -o $(BIN_OUT)/$(BIN_NAME)

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	GOOS=$(@D) GOARCH=$(@F) go build -o $(BIN_OUT)/$(BIN_NAME)-$(@D)-$(@F)

.PHONY: clean
clean:
	rm -rf bin/