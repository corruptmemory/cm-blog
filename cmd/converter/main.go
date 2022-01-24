package main

import (
	"log"
	"os"

	"github.com/corruptmemory/cm-blog/org"
)

func main() {
	log.Println("Hello from the converter!")
	parser, products := org.NewScanner()
	bytes, err := os.ReadFile("content/homepage.org")
	if err != nil {
		log.Fatalf("bad things happened: %s", err)
	}
	stuff := string(bytes)

	stuff1 := stuff[0 : len(stuff)/2]
	stuff2 := stuff[len(stuff1):]

	err = parser.Consume(stuff1)
	if err != nil {
		log.Fatalf("nope1: %s", err)
	}
	err = parser.Consume(stuff2)
	if err != nil {
		log.Fatalf("nope2: %s", err)
	}
	parser.EOF()
	for r := range products {
		log.Println(r)
	}
}
