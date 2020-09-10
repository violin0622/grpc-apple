#!/bin/sh
protoc \
    --go_out=paths=source_relative:.  \
    api/*.proto

protoc \
    -I. \
    -I$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.8/third_party/googleapis    \
    --go_out=paths=source_relative:. \
    --go-grpc_out=paths=source_relative:. \
    --grpc-gateway_out=paths=source_relative:. \
    api/operation/*.proto
