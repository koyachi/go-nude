package main

import (
	"fmt"
	"log"

	"github.com/koyachi/go-nude"
)

func main() {
	//imagePath := "../images/damita.jpg"
	//imagePath := "../images/damita2.jpg"
	imagePath := "../images/test2.jpg"
	//imagePath := "../images/test6.jpg"

	isNude, err := nude.IsNude(imagePath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("isNude = %v\n", isNude)
}
