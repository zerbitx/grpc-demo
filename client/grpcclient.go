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

type (
	// Client wraps proto.GuestBookServiceClient
	Client struct {
		service proto.GuestBookServiceClient
	}

	// Middleware is used to chain grpc.UnaryInvokers in a decorator pattern.
	Middleware func(grpc.UnaryInvoker) grpc.UnaryInvoker
)

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

// Chain wraps a variadic set of grpc.UnaryInvoker Middleware
func Chain(outer Middleware, mws ...Middleware) Middleware {
	return func(next grpc.UnaryInvoker) grpc.UnaryInvoker {
		for i := len(mws) - 1; i >= 0; i-- {
			next = mws[i](next)
		}
		return outer(next)
	}
}

// Interceptor adds logging and deadline Middleware to every request
func Interceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	inv := Chain(TimerMiddleware, DeadlineMiddleware)(invoker)

	err := inv(ctx, method, req, reply, cc, opts...)
	if err != nil {
		return err
	}

	return nil
}

// TimerMiddleware adds time & success logging around a grpc.UnaryInvoker
func TimerMiddleware(next grpc.UnaryInvoker) grpc.UnaryInvoker {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		var err error
		defer func(start time.Time) {
			status := "Success"
			if err != nil {
				status = "Fail"
			}

			fmt.Printf("%s: [%s] took=[%s]\n", status, method, time.Since(start))
		}(time.Now())

		err = next(ctx, method, req, reply, cc, opts...)
		return err
	}
}

// DeadlineMiddleware wraps method logging and the addtion of a context deadline if one
// is not present.
func DeadlineMiddleware(next grpc.UnaryInvoker) grpc.UnaryInvoker {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		fmt.Printf("Method: [%s]\n", method)
		// add a deadline if none is specified
		_, ok := ctx.Deadline()
		if !ok {
			var done func()
			ctx, done = context.WithTimeout(ctx, 1*time.Millisecond)
			defer done()
		}

		return next(ctx, method, req, reply, cc, opts...)
	}
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
