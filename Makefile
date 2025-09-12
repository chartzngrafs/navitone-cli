APP_NAME := navitone
BIN_DIR := bin
CMD_DIR := ./cmd/$(APP_NAME)

.PHONY: build run install tidy clean

build:
	@echo "Building $(APP_NAME) -> $(BIN_DIR)/$(APP_NAME)"
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)

run:
	go run $(CMD_DIR)

install:
	go install $(CMD_DIR)

tidy:
	go mod tidy

clean:
	rm -f $(BIN_DIR)/$(APP_NAME)

