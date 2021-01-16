package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/goccy/go-graphviz"
)

func renderGraph(input io.Reader, output io.Writer) error {
	b, err := ioutil.ReadAll(input)
	if err != nil {
		return err
	}
	fmt.Println(string(b))

	graph, err := graphviz.ParseBytes(b)
	if err != nil {
		return err
	}
	// Hack to fix empty names when there is a label
	node := graph.FirstNode()
	for node != nil {
		fmt.Printf("label: %s and name: %s\n", node.Get("label"), node.Name())
		label := node.Get("label")
		if label == "" {
			node.Set("label", node.Name())
		}
		node = graph.NextNode(node)
	}
	g := graphviz.New()

	if err := g.Render(graph, graphviz.PNG, output); err != nil {
		return err
	}
	return nil
}
func usage() {
	fmt.Println(`
Usaage: dotdoc action
action: serve | render
`)
}
func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}
	fmt.Printf("Args %v\n", os.Args)
	action := os.Args[1]
	switch action {
	case "serve":
		serve()
	case "render":
		render()
	default:
		usage()
	}
}
func render() {

	var input io.Reader
	var output io.Writer
	if len(os.Args) == 2 {
		input = os.Stdin
		output = os.Stdout
	}
	var err error
	if len(os.Args) > 2 {
		inp := os.Args[2]
		input, err = os.Open(inp)
		if err != nil {
			log.Fatal(err)
		}
	}
	if len(os.Args) > 3 {
		outp := os.Args[3]
		output, err = os.Create(outp)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = renderGraph(input, output)

	if err != nil {
		log.Fatal(err)
	}

}
