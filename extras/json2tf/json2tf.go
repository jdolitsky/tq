package main

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/jdolitsky/tq/pkg/cli"
	"github.com/jdolitsky/tq/pkg/tq"
)

func main() {
	buffer, _, err := cli.CommandLineArgsBuffer(false)
	if err != nil {
		log.Fatal(err)
	}
	file, err := tq.ParseJSON(buffer.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(hclwrite.Format(file.Bytes())))
}
