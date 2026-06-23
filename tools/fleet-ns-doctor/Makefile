.PHONY: build test validate plan

build:
	go build ./cmd/fleet-ns-doctor

test:
	go test ./...

validate:
	go run ./cmd/fleet-ns-doctor validate

plan:
	go run ./cmd/fleet-ns-doctor plan
