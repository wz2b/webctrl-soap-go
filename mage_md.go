//go:build mage
// +build mage

package main

import (
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"io/ioutil"
	"log"
)

func MdTest() {
	input, err := ioutil.ReadFile("docs/trend_export.md")
	if err != nil {
		log.Fatalf("Unable to read file: %s\n", err)
	}

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	output := markdown.ToHTML(input, nil, renderer)

	fmt.Println(string(output))

}
