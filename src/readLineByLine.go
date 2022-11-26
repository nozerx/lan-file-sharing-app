package src

import (
	"bufio"
	"fmt"
	"os"
)

func PrintLineByLine(name string) {
	file, err := os.Open(name)
	if err != nil {
		fmt.Println("Error while opening the file", name)
		panic(err)
	} else {
		fmt.Println("successfully opened the file", name)
	}
	reader := bufio.NewReader(file)
	exit := false
	for !exit {
		line, _ := reader.ReadString('\n')
		if line == "" {
			exit = true
		}
		fmt.Println(line)
	}

}
