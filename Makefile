APP=lenta-scraper
PKG=./cmd/$(APP)
CONFIG=configs/local.yaml
OUT=dump.csv

.PHONY: tidy build run clean fmt vet

tidy:
	go mod tidy

build:
	go build -o bin/$(APP) $(PKG)

run:
	go run $(PKG) --config $(CONFIG)

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -rf bin
	rm -f $(OUT)
