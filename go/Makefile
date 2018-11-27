.PHONY: install build clean cleanall

# See https://grpc.io/docs/quickstart/go.html.
install:
	go get -u google.golang.org/grpc
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get -u github.com/mitchellh/go-homedir
	go get -u github.com/spf13/cobra
	go get -u github.com/spf13/viper

build:
	protoc -I ../ ../kv.proto --go_out=plugins=grpc:pkg/kv

clean:

cleanall: clean
	rm -f ./pkg/kv/kv.pb.go