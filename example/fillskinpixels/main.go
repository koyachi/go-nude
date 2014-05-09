package main

import (
	"../../"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
)

func fillSkinPixels(imagePath string, regions nude.Regions) {
	path, err := filepath.Abs(imagePath)
	if err != nil {
		log.Fatal(err)
	}

	img, err := nude.DecodeImage(path)
	dstImg := image.NewRGBA(img.Bounds())
	draw.Draw(dstImg, img.Bounds(), img, image.ZP, draw.Src)

	for i := 0; i < len(regions); i++ {
		blue := color.RGBA{0, 0, 255, 255}
		for _, pixel := range regions[i] {
			dstImg.Set(pixel.X, pixel.Y, blue)
		}
	}

	path, err = filepath.Abs("result.jpg")
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	err = jpeg.Encode(file, dstImg, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	imagePath := "../images/damita.jpg"
	//imagePath := "../images/damita2.jpg"
	//imagePath := "../images/test2.jpg"
	//imagePath := "../images/test6.jpg"

	d := nude.NewDetector(imagePath)
	isNude, err := d.Parse()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("isNude = %v\n", isNude)
	fmt.Printf("%s\n", d)
	fillSkinPixels(imagePath, d.SkinRegions)
}
