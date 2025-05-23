// Package proto contains all the generated Go
// code based on protocol-buffers for gRPC.
package proto

//go:generate protoc --proto_path=. --proto_path=../third_party --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative thrippy/v1/oauth.proto
//go:generate protoc --proto_path=. --proto_path=../third_party --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative thrippy/v1/thrippy.proto
