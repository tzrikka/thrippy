.PHONY: go
go:
	go generate ./proto
	go build ./cmd/thrippy

# https://protobuf.dev/installation/
# https://grpc.io/docs/languages/go/quickstart/
.PHONY: protoc-plugins
protoc-plugins:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# https://github.com/googleapis/googleapis/tree/master/google/api
.PHONY: third-party-deps
third-party-deps:
	curl --create-dirs https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto    -o third_party/google/api/annotations.proto
	curl --create-dirs https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/client.proto         -o third_party/google/api/client.proto
	curl --create-dirs https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/field_behavior.proto -o third_party/google/api/field_behavior.proto
	curl --create-dirs https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto           -o third_party/google/api/http.proto

all: protoc-plugins third-party-deps go

.PHONY: clean
clean:
	rm proto/thrippy/v1/*.pb.go
	rm -rf third_party
	rm thrippy
