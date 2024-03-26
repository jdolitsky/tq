package main

import (
	"fmt"
	"log"

	"github.com/jdolitsky/tq/pkg/cli"
)

func main() {
	buf, query, err := cli.CommandLineArgsBuffer(true)
	if err != nil {
		log.Fatal(err)
	}
	tqb, err := cli.Query(query, buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(tqb))
}
