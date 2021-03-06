.PHONY: all
all: clean test build

.PHONY: clean
clean:
	rm -rf ./dist

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint:
	CGO_ENABLED=0 golangci-lint run --enable-all

.PHONY: test
test: pkg/types/build_types.pb.go
	CGO_ENABLED=0 go test -cover ./...

.PHONY: build
build: pkg/types/build_types.pb.go
	goreleaser build --snapshot --rm-dist

%.pb.go: %.proto
	protoc --go_out=plugins=grpc:. --go_opt=module=github.com/dodo-cli/dodo-build $<
