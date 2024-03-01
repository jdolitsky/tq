package cli

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
)

// Inspired by https://github.com/tmccombs/hcl2json/blob/main/main.go
// Allow for input to come from STDIN or multiple files provided
// as program arguments. If withQuery is true, assume first arg
// is the jq query to be used
func CommandLineArgsBuffer(withQuery bool) (*bytes.Buffer, string, error) {
	flag.Parse()
	files := flag.Args()
	query := ""
	if withQuery {
		if len(files) == 0 {
			return nil, "", fmt.Errorf("missing jq query")
		}
		query = files[0]
		files = files[1:]
	}
	var inputName string
	switch len(files) {
	case 0:
		files = append(files, "-")
		inputName = "STDIN"
	case 1:
		inputName = files[0]
		if inputName == "-" {
			inputName = "STDIN"
		}
	default:
		inputName = "COMPOSITE"
	}
	buffer := bytes.NewBuffer([]byte{})
	for _, filename := range files {
		var stream io.Reader
		if filename == "-" {
			stream = os.Stdin
			filename = "STDIN" // for better error message
		} else {
			file, err := os.Open(filename)
			if err != nil {
				return nil, "", fmt.Errorf("Failed to open %s: %s\n", filename, err)
			}
			defer file.Close()
			stream = file
		}
		_, err := buffer.ReadFrom(stream)
		if err != nil {
			return nil, "", fmt.Errorf("Failed to read from %s: %s\n", filename, err)
		}
		buffer.WriteByte('\n') // just in case it doesn't have an ending newline
	}
	return buffer, query, nil
}
