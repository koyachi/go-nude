package nude

import (
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
)

// experimental
func DecodeImage(filePath string) (img image.Image, err error) {
	return decodeImage(filePath)
}

func decodeImage(filePath string) (img image.Image, err error) {
	reader, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	last3Strings := strings.ToLower(filePath[len(filePath)-3:])
	last4Strings := strings.ToLower(filePath[len(filePath)-4:])
	if last3Strings == "jpg" || last4Strings == "jpeg" {
		img, err = jpeg.Decode(reader)
	} else if last3Strings == "gif" {
		img, err = gif.Decode(reader)
	} else if last3Strings == "png" {
		img, err = png.Decode(reader)
	} else {
		img = nil
		err = errors.New("unknown format")
	}
	return
}
