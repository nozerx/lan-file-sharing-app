package main

import (
	"lanfilesharing/src"
)

func main() {
	src.PrintLineByLine("src/readLineByLine.go")
	src.ReadInBinaryFromFile("main.go")
}
