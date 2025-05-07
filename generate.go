package trippy

//go:generate protoc --proto_path=proto --proto_path=third_party --go_out=proto --go-grpc_out=proto --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative trippy/v1/oauth_config.proto
//go:generate protoc --proto_path=proto --proto_path=third_party --go_out=proto --go-grpc_out=proto --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative trippy/v1/trippy.proto
