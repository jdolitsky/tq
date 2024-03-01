package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/jdolitsky/tq/pkg/cli"
	"github.com/jdolitsky/tq/pkg/tq"
)

func main() {
	buffer, _, err := cli.CommandLineArgsBuffer(false)
	if err != nil {
		log.Fatal(err)
	}
	tfFile, err := tq.ParseTerraform(buffer.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	jb, err := json.Marshal(tfFile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(jb))
}
