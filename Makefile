.PHONY: build run test

APP=chapteriser
BIN_DIR=./bin
BIN=$(BIN_DIR)/$(APP)

log = @printf "\033[1;32m ►\033[0m %s\n" "$(1)"

build:
	$(call log,Building...)
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN) .
	$(call log,Build done!)

run:
	@$(BIN) -h

test:
	$(call log,Testing...)
	@go test ./...

clean:
	$(call log,Cleaning...)
	@rm -rf $(BIN_DIR)
	$(call log,Removed $(BIN_DIR))
