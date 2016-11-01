package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {

	fmt.Println("Connecting to gRPC server")

	c, err := NewClient("127.0.0.1:8888")
	if err != nil {
		fmt.Printf("Unable to connect: %s\n", err)
		os.Exit(1)
	}

	for {

		w := command()

		switch w {
		case "l":
			err = c.List()
		case "c":
			err = c.Create()
		default:
			fmt.Printf("Unknown command [%s]\n", w)
		}

		if err != nil {
			fmt.Println("grpc error: %s\n", err)
			os.Exit(1)
		}
	}

}

func command() string {

	r := bufio.NewReader(os.Stdin)

	fmt.Print("List (l) or Create (c): ")
	c, _ := r.ReadString('\n')

	return strings.ToLower(strings.TrimSpace(c))
}
