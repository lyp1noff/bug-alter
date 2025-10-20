APP_NAME=sea-battle
ENTRYPOINT=.
OUTPUT_DIR=bin

.PHONY: run fmt build build-linux build-windows clean

run:
	go run $(ENTRYPOINT)

fmt:
	go fmt ./...

build:
	mkdir -p $(OUTPUT_DIR)
	go build -o $(OUTPUT_DIR)/$(APP_NAME) $(ENTRYPOINT)

build-linux:
	mkdir -p $(OUTPUT_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(OUTPUT_DIR)/$(APP_NAME)-linux $(ENTRYPOINT)

build-windows:
	mkdir -p $(OUTPUT_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(OUTPUT_DIR)/$(APP_NAME).exe $(ENTRYPOINT)

clean:
	rm -rf $(OUTPUT_DIR)
