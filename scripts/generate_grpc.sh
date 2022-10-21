#!/bin/bash
protoc --go_out=internal/grpc/ --go_opt=paths=source_relative \
    --go-grpc_out=internal/grpc/ --go-grpc_opt=paths=source_relative \
    -Iinternal/grpc/ \
    internal/grpc/error.proto

protoc --go_out=internal/grpc/ --go_opt=paths=source_relative \
    --go-grpc_out=internal/grpc/ --go-grpc_opt=paths=source_relative \
    -Iinternal/grpc/ \
    internal/grpc/asset_transfer.proto