package cli

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
)

func TestQuery(t *testing.T) {
	// Note: when adding tests here, be mindful of tabs (\t) vs. spaces
	tests := []struct {
		name      string
		query     string
		input     string
		want      string
		shouldErr bool
	}{
		{
			name:  "returns self",
			query: `.`,
			input: dedent.Dedent(`
				resource "oci_tag" "latest" {
				  depends_on = [module.test-latest]
				  digest_ref = module.latest.image_ref
				  tag        = "latest"
				}
			`),
			want: dedent.Dedent(`
				resource "oci_tag" "latest" {
				  depends_on = [module.test-latest]
				  digest_ref = module.latest.image_ref
				  tag        = "latest"
				}
		    `),
		},
		{
			name:  "extract inner value, string",
			query: `.body.blocks[] | select(.type == "resource" and .labels[0] == "oci_tag" and .labels[1] == "latest-dev").attributes["tag"]`,
			input: dedent.Dedent(`
				resource "other" {
				  a = 1
				}
				
				resource "oci_tag" "latest" {
				  depends_on = [module.test-latest]
				  digest_ref = module.latest.image_ref
				  tag        = "latest"
				}

				resource "oci_tag" "latest-dev" {
				  depends_on = [module.test-latest]
				  digest_ref = module.latest.dev_ref
				  tag        = "latest-dev"
				}
			`),
			want: `"latest-dev"`,
		},
		{
			name:  "extract inner value, non-string",
			query: `.body.blocks[] | select(.type == "resource" and .labels[0] == "oci_tag" and .labels[1] == "latest-dev").attributes["digest_ref"]`,
			input: dedent.Dedent(`
				resource "other" {
				  a = 1
				}
				
				resource "oci_tag" "latest" {
				  depends_on = [module.test-latest]
				  digest_ref = module.latest.image_ref
				  tag        = "latest"
				}

				resource "oci_tag" "latest-dev" {
				  depends_on = [module.test-latest]
				  digest_ref = module.latest.dev_ref
				  tag        = "latest-dev"
				}
			`),
			want: "module.latest.dev_ref",
		},
		{
			name:  "delete entire sections",
			query: `del(.body.blocks[] | select(.type == "terraform" or (.type == "resource" and .labels[0] == "oci_tag")))`,
			input: dedent.Dedent(`
				terraform {
				  required_providers {
				    oci = { source = "chainguard-dev/oci" }
				  }
				}
				resource "other" {
				  a = 1
				}
				  
				resource "oci_tag" "latest" {
				  depends_on = [module.test-latest]
				  digest_ref = module.latest.image_ref
				  tag        = "latest"
				}
  
				resource "oci_tag" "latest-dev" {
				  depends_on = [module.test-latest]
				  digest_ref = module.latest.dev_ref
				  tag        = "latest-dev"
				}
			`),
			want: dedent.Dedent(`
				resource "other" {
				  a = 1
				}
			`),
		},
		{
			name:  "append attribute",
			query: `.body.blocks[] |= if (.type == "resource" and .labels[0] == "other") then (.attributes["b"] = "2") else . end`,
			input: dedent.Dedent(`
				resource "oci_tag" "latest" {
				  depends_on = [module.test-latest]
				  digest_ref = module.latest.image_ref
				  tag        = "latest"
				}

				resource "other" {
				  a = 1
				}
			`),
			want: dedent.Dedent(`
				resource "oci_tag" "latest" {
				  depends_on = [module.test-latest]
				  digest_ref = module.latest.image_ref
				  tag        = "latest"
				}

				resource "other" {
				  a = 1
				  b = 2
				}
			`),
		},
		{
			name:  "append entire block",
			query: `.body.blocks += [{type: "resource", labels: ["testing"], attributes: {"hello": "\"world\""}}]`,
			input: dedent.Dedent(`
				resource "other" {
				  a = 1
				}
			`),
			want: dedent.Dedent(`
				resource "other" {
				  a = 1
				}

				resource "testing" {
				  hello = "world"
				}
			`),
		},
		{
			name:  "extract from locals",
			query: `.body.blocks[] | select(.type == "locals").attributes["versions"]`,
			input: dedent.Dedent(`
				terraform {
				  required_providers {
				    oci = { source = "chainguard-dev/oci" }
				  }
				}
				locals {
				  versions = ["1.2.3", "4.5.6"]
				}
			`),
			want: `["1.2.3", "4.5.6"]`,
		},
		{
			name:  "nested block extraction",
			query: `.body.blocks[] | select(.type == "resource" and .labels[0] == "cosign_attest").body.blocks[] | select(.type == "predicates").attributes["type"]`,
			input: dedent.Dedent(`
				terraform {
				  required_providers {
				    cosign = {
				      source = "chainguard-dev/cosign"
				    }
				  }
				}
				
				resource "cosign_attest" "yolo" {
				  conflict = "SKIPSAME"
				  image    = module.image-yolo["yolo"].image_ref
				  predicates {
				    json = file("${path.module}/yolo.json")
				    type = "https://yolo.example.com/ns"
				  }
				}
			`),
			want: `"https://yolo.example.com/ns"`,
		},
		{
			name:  "bad terraform",
			query: `.`,
			input: dedent.Dedent(`
				terraform {
				  required_providers {
			`),
			shouldErr: true,
		},
		{
			name:  "bad jq query",
			query: `YOLO .. .. .YOOOOOO`,

			input: dedent.Dedent(`
				terraform {
				  required_providers {
				    cosign = {
				      source = "chainguard-dev/cosign"
				    }
				  }
				}
			`),
			shouldErr: true,
		},
	}

	for _, test := range tests {
		got, err := Query(test.query, []byte(test.input))
		if err != nil {
			if test.shouldErr {
				continue
			}
			t.Errorf("%s: TQ() returned error: %v", test.name, err)
		}
		if diff := cmp.Diff(strings.TrimSpace(string(got)), strings.TrimSpace(test.want)); diff != "" {
			t.Errorf("%s: did not get expected output %s", test.name, diff)
		}
	}
}
