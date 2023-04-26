package colorutils

import (
	"errors"
	"image"
	"math"

	"github.com/zekurio/daemon/pkg/httputils"
)

// GenerateColorFromImage generates a HEX color from an images prominent color.
func GenerateColorFromImage(img image.Image) (color int, err error) {

	// define our bounds and color vars
	bounds := img.Bounds()
	var r, g, b uint32
	var count uint32

	// loop through the image and get the average color
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			pr, pg, pb, _ := img.At(x, y).RGBA()
			r += pr
			g += pg
			b += pb
			count++
		}
	}

	// calculate the average color
	r /= count
	g /= count
	b /= count

	// convert the color to a hex value
	color = int(r>>8)<<16 | int(g>>8)<<8 | int(b>>8)

	return
}

// GenerateColorFromImageURL generates a HEX color from an images prominent color from a URL.
func GenerateColorFromImageURL(url string) (color int, err error) {

	var img image.Image

	data, contentType, err := httputils.GetFile(url, nil)
	if err != nil {
		return
	} else if contentType != "image/jpeg" && contentType != "image/png" {
		return math.MaxInt, errors.New("invalid image type")
	}

	img, _, err = image.Decode(data)
	if err != nil {
		return
	}

	return GenerateColorFromImage(img)

}
