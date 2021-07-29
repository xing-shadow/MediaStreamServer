SHELL=/bin/bash

.PHONY: build run clean

build:
	go build -o RtspServer  main.go

run:
	go build -o RtspServer  main.go
	./RtspServer -c ./config/Config.toml

clear:
	rm RtspServer -rf