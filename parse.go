package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"github.com/russross/blackfriday/v2"
)

var imageCache [][]byte

func init() {
	imageCache = make([][]byte, 0)
}

func addImage(img []byte) int {
	i := len(imageCache)
	imageCache = append(imageCache, img)
	return i
}
func processCode(bits []byte) ([]byte, error) {
	buf := bufio.NewReader(bytes.NewReader(bits))
	l, _, err := buf.ReadLine()
	if err != nil {
		return nil, err
	}
	line := string(l)

	// If starts with graph file
	if strings.Index(line, "dot") != 0 && strings.Index(line, "graphviz") != 0 {
		return nil, nil
	}
	output := bytes.NewBuffer([]byte{})
	err = renderGraph(buf, output)
	if err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
func parseMarkdown(input io.Reader) ([]byte, error) {

	b, err := ioutil.ReadAll(input)
	if err != nil {
		return nil, err
	}
	md := blackfriday.New()
	root := md.Parse(b)
	root.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {

		if node == nil {
			return blackfriday.Terminate
		}
		if node.Type == blackfriday.CodeBlock {
			fmt.Println("Yeah it is a code block")
			fmt.Printf(string(node.CodeBlockData.Info))
		}
		if node.Type == blackfriday.Code {
			fmt.Println("Yeah it is a code")
			fmt.Printf(string(node.Literal))
			img, err := processCode(node.Literal)
			if err != nil {
				log.Printf("Error processing code: %s\n", err)
			}
			if img != nil {

				imageId := addImage(img)
				fmt.Printf("Image added %d\n", imageId)
				path := fmt.Sprintf("/_grok/img/%d.png", imageId)
				fmt.Printf("Path is %s\n", path)
				newNode := blackfriday.NewNode(blackfriday.Image)
				newNode.LinkData = blackfriday.LinkData{Destination: []byte(path)}
				newNode.Prev = node
				newNode.Next = node.Next
				node.Next = newNode
				//node.InsertBefore(newNode)
			}
		}

		return blackfriday.GoToNext
	})
	r := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags,
	})
	var buf bytes.Buffer

	r.RenderHeader(&buf, root)
	root.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		return r.RenderNode(&buf, node, entering)
	})
	r.RenderFooter(&buf, root)
	return buf.Bytes(), nil
}
