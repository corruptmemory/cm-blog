package main

import (
	"log"
	"os"

	"github.com/corruptmemory/cm-blog/org"
)

func main() {
	log.Println("Hello from the converter!")
	bytes, err := os.ReadFile("content/homepage.org")
	if err != nil {
		log.Fatalf("bad things happened: %s", err)
	}
	parser, products := org.NewScanner(string(bytes))
	_ = parser.Scan()
	for r := range products {
		log.Println(r)
	}
}
