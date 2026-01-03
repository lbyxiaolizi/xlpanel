BIN_DIR := bin

.PHONY: all server emailpipe mock_plugin

all: server emailpipe mock_plugin

server:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/server ./cmd/server

emailpipe:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/emailpipe ./cmd/emailpipe

mock_plugin:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/mock_plugin ./cmd/mock_plugin
