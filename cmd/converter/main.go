package main

import (
	"log"
	"os"

	"github.com/corruptmemory/cm-blog/org"
)

func main() {
	log.Println("Hello from the converter!")
	parser, products := org.NewParser()
	bytes, err := os.ReadFile("content/homepage.org")
	if err != nil {
		log.Fatalf("bad things happened: %s", err)
	}
	stuff := string(bytes)

	err = parser.Consume(stuff)
	parser.EOF()
	for r := range products {
		log.Println(r)
	}
}
