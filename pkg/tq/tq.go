package tq

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type (
	TerraformFile struct {
		Body TerraformFileBody `json:"body"`
	}

	TerraformFileBody struct {
		Blocks []TerraformFileBlock `json:"blocks"`
	}

	TerraformFileBlock struct {
		Type       string            `json:"type"`
		Labels     []string          `json:"labels"`
		Attributes map[string]string `json:"attributes"`
		Body       TerraformFileBody `json:"body"`
	}

	HCLWriteFile struct {
		*hclwrite.File
	}
)

func ParseTerraform(b []byte) (*TerraformFile, error) {
	file, diags := hclwrite.ParseConfig(b, "", hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf(diags.Error())
	}
	return Serialize(&HCLWriteFile{file}), nil
}

func Serialize(file *HCLWriteFile) *TerraformFile {
	tfFile := TerraformFile{
		Body: tfBodyToBody(file.Body()),
	}
	return &tfFile
}

func ParseJSON(b []byte) (*HCLWriteFile, error) {
	var tfFile TerraformFile
	if err := json.Unmarshal(b, &tfFile); err != nil {
		return nil, err
	}
	return Deserialize(&tfFile), nil
}

func Deserialize(tfFile *TerraformFile) *HCLWriteFile {
	file := hclwrite.NewEmptyFile()
	transferBodyToHCLBody(tfFile.Body, file.Body(), 0)
	return &HCLWriteFile{file}
}

func (tfFile *TerraformFile) String() string {
	return strings.TrimSuffix(string(hclwrite.Format(Deserialize(tfFile).Bytes())), "\n")
}

func (hclWriteFile *HCLWriteFile) String() string {
	return strings.TrimSuffix(string(hclwrite.Format(hclWriteFile.Bytes())), "\n")
}

func tfBodyToBody(tfBody *hclwrite.Body) TerraformFileBody {
	body := TerraformFileBody{
		Blocks: []TerraformFileBlock{},
	}
	for _, block := range tfBody.Blocks() {
		attributes := map[string]string{}
		// Note: the .Attributes() call here returns a map
		// in random order, so there isn't a way to really retain
		// the order in the terraform file...
		for k, v := range block.Body().Attributes() {
			attributes[k] = strings.TrimLeft(string(v.Expr().BuildTokens(nil).Bytes()), " ")
		}
		body.Blocks = append(body.Blocks, TerraformFileBlock{
			Type:       block.Type(),
			Labels:     block.Labels(),
			Attributes: attributes,
			Body:       tfBodyToBody(block.Body()),
		})
	}
	return body
}

func transferBodyToHCLBody(body TerraformFileBody, hclBody *hclwrite.Body, level int) {
	for _, tqBlock := range body.Blocks {
		hclBlock := hclwrite.NewBlock(tqBlock.Type, tqBlock.Labels)
		// sort the attribute keys so we hav deterministic output
		keys := make([]string, 0, len(tqBlock.Attributes))
		for k := range tqBlock.Attributes {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			hclBlock.Body().SetAttributeRaw(k, hclwrite.Tokens{{
				Bytes: []byte(" " + tqBlock.Attributes[k]),
			}})
		}
		transferBodyToHCLBody(tqBlock.Body, hclBlock.Body(), level+1)
		hclBody.AppendBlock(hclBlock)
		if level == 0 {
			// Only append newlines for top-level blocks
			hclBody.AppendNewline()
		}
	}
}
