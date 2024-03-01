package tq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
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
)

// TQ does the following:
// 1. Converts incoming Terraform bytes to JSON
// 2. Runs it through jq (via subprocess) to apply the query
// 3. Converts it back to Terraform (Note: if this fails, just return the JSON)
func TQ(query string, in []byte) ([]byte, error) {
	// Convert original input to JSON
	tfFile, err := ParseTerraform(in)
	if err != nil {
		return nil, err
	}
	jb, err := json.Marshal(tfFile)
	if err != nil {
		return nil, err
	}

	// Run jq directly, passing the parsed JSON as STDIN
	cmd := exec.Command("jq", "-r", query)
	var jqbuff bytes.Buffer
	cmd.Stdout = &jqbuff
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if _, err := io.WriteString(stdin, string(jb)); err != nil {
		return nil, err
	}
	stdin.Close()
	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	// Now attempt to convert it back to tf
	jqb := jqbuff.Bytes()
	file, err := ParseJSON(jqb)
	if err != nil {
		// TODO: add option to fail here
		// If there was some issue converting back to terraform,
		// just print out the JSON (assume what the user wants here)
		return jqb, nil
	}
	tfb := hclwrite.Format(file.Bytes())
	if tfb == nil {
		// TODO: add option to fail here
		// Like above, maybe we somehow passed conversion, but
		// the result is empty. Return the jq raw output.
		return jqb, nil
	}

	return tfb, nil
}

func ParseTerraform(b []byte) (*TerraformFile, error) {
	file, diags := hclwrite.ParseConfig(b, "", hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf(diags.Error())
	}
	return Serialize(file), nil
}

func Serialize(file *hclwrite.File) *TerraformFile {
	tfFile := TerraformFile{
		Body: tfBodyToBody(file.Body()),
	}
	return &tfFile
}

func ParseJSON(b []byte) (*hclwrite.File, error) {
	var tfFile TerraformFile
	if err := json.Unmarshal(b, &tfFile); err != nil {
		return nil, err
	}
	return Deserialize(&tfFile), nil
}

func Deserialize(tfFile *TerraformFile) *hclwrite.File {
	file := hclwrite.NewEmptyFile()
	transferBodyToHCLBody(tfFile.Body, file.Body(), 0)
	return file
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
