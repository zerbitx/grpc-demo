package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/weave-lab/grpc-demo/grpcdemoproto"
)

func main() {
	// Create a listener to accept incoming requests
	listener, err := net.Listen("tcp", ":8888")
	if err != nil {
		os.Exit(1)
	}

	// Create a gRPC server with a logging middleware
	server := grpc.NewServer(grpc.UnaryInterceptor(UnaryLogging))

	// Register our GuestBookService implementation with the server
	grpcdemoproto.RegisterGuestBookServiceServer(server, NewGuestBookService())

	fmt.Println("Serving on", listener.Addr().String())
	server.Serve(listener)
}

//GuestBookService implements the gRPC service defined in the proto file
type GuestBookService struct {
	guestBookEntries []*grpcdemoproto.GuestBookEntry
}

//Create adds a new entry to the guestbook
func (svc *GuestBookService) Create(ctx context.Context, entry *grpcdemoproto.GuestBookEntry) (*empty.Empty, error) {
	now, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return nil, err
	}
	entry.Time = now

	svc.guestBookEntries = append(svc.guestBookEntries, entry)

	return &empty.Empty{}, nil
}

//List lists all current entries in the guestbook
func (svc *GuestBookService) List(ctx context.Context, _ *empty.Empty) (*grpcdemoproto.ListGuestBookResponse, error) {
	return &grpcdemoproto.ListGuestBookResponse{
		Entries: svc.guestBookEntries,
	}, nil
}

//NewGuestBookService creates a GuestBookService instance with some example entries
func NewGuestBookService() *GuestBookService {
	now, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		os.Exit(1)
	}

	return &GuestBookService{
		guestBookEntries: []*grpcdemoproto.GuestBookEntry{
			&grpcdemoproto.GuestBookEntry{
				Name:    "Robison Rogers",
				Message: "Dammit Clint.",
				Time:    now,
			},
			&grpcdemoproto.GuestBookEntry{
				Name:    "Colton Shields",
				Message: "I like to leave early on Fridays... If I come in at all.",
				Time:    now,
			},
		},
	}
}

//UnaryLogging is a gRPC interceptor for logging simple messages when requests are received
func UnaryLogging(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	fmt.Printf("Received gRPC Request for method [%s] at [%s]\n", info.FullMethod, time.Now())

	resp, err = handler(ctx, req)

	fmt.Printf("Finished handling gRPC request at [%s]\n", time.Now())

	return resp, err
}
