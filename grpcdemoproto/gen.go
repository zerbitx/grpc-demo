package grpcdemoproto

//go:generate protoc -I=./ -I=$GOPATH/src/github.com/weave-lab/grpc-demo/vendor --go_out=plugins=grpc:$GOPATH/src grpcdemo.proto
