protoc --proto_path=. --go_out=..\datacloak\server --go-grpc_out=..\datacloak\server --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative .\rpc_meta.proto .\hello_world.proto
