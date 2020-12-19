package main

import (
	"context"
	"fmt"
	"log"

	"github.com/xabi93/aimharder/example"
)

func main() {
	cli := example.Client()
	ctx := context.Background()

	me, err := cli.Users.Me(ctx)
	if err != nil {
		log.Fatal(err)
	}

	bookings, err := cli.Bookings.Next(ctx, me.BoxID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", bookings)
}
