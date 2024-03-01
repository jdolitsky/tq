package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/jdolitsky/tq/pkg/tq"
)

func main() {
	tfb, err := os.ReadFile("./examples/aws.tf")
	if err != nil {
		log.Fatal(err)
	}

	tfFile, err := tq.ParseTerraform(tfb)
	if err != nil {
		log.Fatal(err)
	}

	// For all "aws_instance" resources,
	// modify the instance_type to "t2.large"
	for _, block := range tfFile.Body.Blocks {
		if block.Type == "resource" {
			if len(block.Labels) > 0 && block.Labels[0] == "aws_instance" {
				block.Attributes["instance_type"] = `"t2.large"`
			}
		}
	}

	// Convert back to Terraform
	file := tq.Deserialize(tfFile)

	// Pretty-print the Terraform
	fmt.Print(string(hclwrite.Format(file.Bytes())))
}
