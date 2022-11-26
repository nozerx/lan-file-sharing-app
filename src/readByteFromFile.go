package src

import (
	"fmt"
	"io"
	"os"
)

func ReadInBinaryFromFile(name string) {
	file, err := os.Open(name)
	if err != nil {
		fmt.Println("Error during opening the file")
	} else {
		fmt.Println("Successfully openend the file ", name)
	}
	defer file.Close()
	buffer := make([]byte, 2)
	for {
		_, err = file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error while reading from the file")
		}
		fmt.Print(string(buffer))
	}

}
