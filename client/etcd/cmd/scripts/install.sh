#!/bin/bash -x
set -e

# if use go 1.16, use go install
go get github.com/gogo/protobuf/protoc-gen-gogofast@latest
go get google.golang.org/protobuf/cmd/protoc-gen-go@latest
go get -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
