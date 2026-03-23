VERSION  = $(or $(shell git tag --points-at HEAD | grep -oP 'v\K[0-9.]+'), unknown)
REVISION = $(shell git rev-parse HEAD)

REPOSITORY := github.com/firmus-public/oob_gpu_exporter
LDFLAGS    := -X $(REPOSITORY)/internal/version.Version=$(VERSION)
LDFLAGS    += -X $(REPOSITORY)/internal/version.Revision=$(REVISION)
GOFLAGS    := -ldflags "$(LDFLAGS)"
RUNFLAGS   ?= -config config.yml -verbose

build: build-linux-amd64 build-linux-arm64

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o oob_gpu_exporter-$(VERSION)-linux-amd64 ./cmd/oob_gpu_exporter

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -o oob_gpu_exporter-$(VERSION)-linux-arm64 ./cmd/oob_gpu_exporter

run:
	go run ./cmd/oob_gpu_exporter $(RUNFLAGS)
