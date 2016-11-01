package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/weave-lab/grpc-demo/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Client struct {
	service proto.GuestBookServiceClient
}

// NewClient connects to the guestbook service and returns a client
func NewClient(addr string) (*Client, error) {

	interceptor := grpc.WithUnaryInterceptor(Interceptor)

	block := grpc.WithBlock()
	timeout := grpc.WithTimeout(time.Second * 2)
	userAgent := grpc.WithUserAgent("guestbook client")

	g, err := grpc.Dial(addr, grpc.WithInsecure(), timeout, userAgent, interceptor, block)
	if err != nil {
		return nil, err
	}

	// get the service client
	c := Client{
		service: proto.NewGuestBookServiceClient(g),
	}

	return &c, err
}

// Interceptor adds logging and deadline middleware to every request
func Interceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	fmt.Printf("Method: [%s]\n", method)

	start := time.Now()

	var err error
	defer func() {
		status := "Success"
		if err != nil {
			status = "Fail"
		}

		fmt.Printf("%s: [%s] took=[%s]\n", status, method, time.Since(start))

	}()

	// add a deadline if none is specified
	_, ok := ctx.Deadline()
	if !ok {
		var done func()
		ctx, done = context.WithTimeout(ctx, time.Second)
		defer done()
	}

	err = invoker(ctx, method, req, reply, cc, opts...)
	if err != nil {
		return err
	}

	return nil

}

// List retrieves all of the guestbook entries and displays them
func (c *Client) List() error {

	ctx := context.Background()

	r, err := c.service.List(ctx, &empty.Empty{})
	if err != nil {
		return err
	}

	fmt.Println("\nListing entries\n------------")
	for _, v := range r.Entries {
		t, _ := ptypes.Timestamp(v.Time)
		fmt.Printf("\t %s %- 16s %s\n", t.Format("Jan 02 15:04"), v.Name, v.Message)
	}
	fmt.Printf("\n")

	return nil
}

// Create creates a new guestbook entry
func (c *Client) Create() error {

	ctx := context.Background()

	r := bufio.NewReader(os.Stdin)

	fmt.Print("Name: ")
	name, _ := r.ReadString('\n')

	fmt.Print("Message: ")
	message, _ := r.ReadString('\n')

	name = strings.TrimSpace(name)
	message = strings.TrimSpace(message)

	return c.create(ctx, name, message)
}

func (c *Client) create(ctx context.Context, name, message string) error {

	in := proto.GuestBookEntry{
		Name:    name,
		Message: message,
	}

	_, err := c.service.Create(ctx, &in)
	if err != nil {
		return err
	}

	return nil
}
