package main

import (
	"context"
	"fmt"
	"log"

	"github.com/xabi93/aimharder/example"
)

func main() {
	cli := example.Client()

	me, err := cli.Users.Me(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", me)
}
