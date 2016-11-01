package grpcdemoproto

//go:generate protoc -I=./ -I=$GOPATH/src --go_out=plugins=grpc:$GOPATH/src grpcdemo.proto
