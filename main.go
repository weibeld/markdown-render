package main

import (
	"fmt"
	"gopkg.in/russross/blackfriday.v2"
	"io/ioutil"
)

func main() {
	markdown := blackfriday.New()
	input, _ := ioutil.ReadFile("test.md")
	node := markdown.Parse(input)
	heading := node.FirstChild
	fmt.Println("Type: " + string(heading.Type))
	if heading.Type == blackfriday.Heading {
		fmt.Println("It is a heading")
	}
	fmt.Println(heading.Parent)
	fmt.Println(heading.FirstChild)
	fmt.Println(heading.LastChild)
	fmt.Println(heading.Prev)
	fmt.Println(heading.Next)
	fmt.Println(string(heading.Literal))
	fmt.Printf("%+v\n", heading.HeadingData)
	fmt.Println(heading.String())
	node.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if node.Type == blackfriday.Paragraph {
			if entering {
				fmt.Println("I found a paragraph")
				text := node.FirstChild
				if text.Type == blackfriday.Text {
					fmt.Println(string(text.Literal))
				}
				//fmt.Println(node.FirstChild)
				//fmt.Println(node.LastChild)
			} else {
				fmt.Println("I'm leaving the paragraph")
			}

		}
		return blackfriday.GoToNext
	})

	//html := blackfriday.Run(input)
	//fmt.Println(string(html))
}
