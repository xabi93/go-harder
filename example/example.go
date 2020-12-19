package example

import (
	"context"
	"log"
	"os"

	"github.com/xabi93/aimharder/aimharder"
)

func Client() *aimharder.Client {
	ctx := context.Background()

	var (
		token = os.Getenv("AIMHARDER_TOKEN")
		mail  = os.Getenv("AIMHARDER_MAIL")
		pw    = os.Getenv("AIMHARDER_PW")

		cli *aimharder.Client
		err error
	)

	cliOps := []aimharder.ClientOption{
		aimharder.OptionDebug(true),
	}

	if token != "" {
		cli, err = aimharder.New(token, cliOps...)
	} else {
		cli, err = aimharder.Login(ctx, mail, pw, cliOps...)
	}
	if err != nil {
		log.Fatal(err)
	}

	return cli
}
