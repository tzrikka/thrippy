package trippy

//go:generate protoc --proto_path=proto --proto_path=third_party --go_out=proto --go-grpc_out=proto --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative thrippy/v1/oauth.proto
//go:generate protoc --proto_path=proto --proto_path=third_party --go_out=proto --go-grpc_out=proto --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative thrippy/v1/thrippy.proto
