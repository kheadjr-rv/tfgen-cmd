package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"encoding/json"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/zclconf/go-cty/cty"
)

func main() {

	optional := flag.Bool("optional", false, "include optional inputs")

	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("Expected two args: src and name")
		os.Exit(1)
	}

	tf, _ := tfconfig.LoadModule(args[0])
	name := args[1]

	keys := make([]string, 0)
	for k, v := range tf.Variables {
		if *optional {
			keys = append(keys, k)
			continue
		}
		if v.Required {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	fmt.Printf("%s", module(name, keys))

	for _, key := range keys {
		fmt.Printf("%s", variable(key, tf.Variables[key]))
	}

	for k := range tf.Outputs {
		s := fmt.Sprintf("# module.%s.%s", name, k)
		fmt.Println(s)
	}
}

func module(name string, keys []string) []byte {
	root := hclwrite.NewEmptyFile()

	block := root.Body().AppendNewBlock("module", []string{name})

	block.Body().SetAttributeValue("source", cty.StringVal("path_goes_here"))
	block.Body().SetAttributeValue("version", cty.StringVal("version_goes_here"))
	block.Body().AppendNewline()

	for _, k := range keys {
		block.Body().SetAttributeRaw(k, tokens([]byte(fmt.Sprintf("var.%s", k))))
	}

	root.Body().AppendNewline()

	return root.Bytes()
}

func variable(name string, v *tfconfig.Variable) []byte {
	root := hclwrite.NewEmptyFile()

	block := root.Body().AppendNewBlock("variable", []string{name})

	if v.Description != "" {
		block.Body().SetAttributeValue("description", cty.StringVal(v.Description))
	}

	if v.Type != "" {
		block.Body().SetAttributeRaw("type", tokens([]byte(v.Type)))
	}

	if v.Default != nil {
		value, _ := json.Marshal(v.Default)
		block.Body().SetAttributeRaw("default", tokens(value))
	}

	root.Body().AppendNewline()

	return root.Bytes()
}

func tokens(b []byte) hclwrite.Tokens {
	return hclwrite.Tokens{
		&hclwrite.Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: b,
		},
	}
}
