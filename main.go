package main

import (
	"fmt"
	"log"

	"github.com/jdolitsky/tq/pkg/cli"
	"github.com/jdolitsky/tq/pkg/tq"
)

func main() {
	buf, query, err := cli.CommandLineArgsBuffer(true)
	if err != nil {
		log.Fatal(err)
	}
	tqb, err := tq.TQ(query, buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(tqb))
}
