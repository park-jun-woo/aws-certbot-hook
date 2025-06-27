# https://parkjunwoo.com/microstral/Makefile
BINARY_NAME=certhook
BUILD_PATH=/usr/local/bin

all: build install setup

build:
	go build -o $(BINARY_NAME) cmd/certhook/main.go

install:
	sudo mv $(BINARY_NAME) $(BUILD_PATH)/
	sudo chmod +x $(BUILD_PATH)/$(BINARY_NAME)
	sudo cp acertbot $(BUILD_PATH)/
	sudo chmod +x $(BUILD_PATH)/acertbot

setup:
	sudo apt update
	sudo apt install ca-certificates -y
	sudo update-ca-certificates

clean:
	rm -f $(BINARY_NAME)
