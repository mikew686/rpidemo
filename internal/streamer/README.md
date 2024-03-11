# streamer protobuf and grpc generated code

Example build lines for Ubuntu:
```
sudo apt install protobuf-compiler
GOPATH=$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest # install to $/bin
GOPATH=~ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative streamsamples.proto
```
