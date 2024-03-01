package main

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/jdolitsky/tq/pkg/tq"
)

func main() {
	// Equivalent of:
	//
	// resource "aws_instance" "app_server_a" {
	//   ami           = local.ami
	//   instance_type = "t2.micro"
	// }
	//
	tfFile := tq.TerraformFile{
		Body: tq.TerraformFileBody{
			Blocks: []tq.TerraformFileBlock{
				{
					Type: "resource",
					Labels: []string{
						"aws_instance",
						"app_server_a",
					},
					Attributes: map[string]string{
						"ami":           "local.ami",
						"instance_type": `"t2.micro"`,
					},
				},
			},
		},
	}

	// Convert to the hashicorp/hcl data type
	file := tq.Deserialize(&tfFile)

	// Pretty-print the Terraform
	fmt.Print(string(hclwrite.Format(file.Bytes())))
}
