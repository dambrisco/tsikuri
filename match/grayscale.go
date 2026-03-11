package match

import (
	"image"
	"image/color"
)

// ToGray converts an image to a grayscale float64 matrix for NCC computation.
// Values are normalized to 0.0–1.0 range.
func ToGray(img image.Image) [][]float64 {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	gray := make([][]float64, h)
	for y := range gray {
		gray[y] = make([]float64, w)
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			// ITU-R BT.601 luminance formula
			lum := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
			gray[y][x] = lum / 65535.0
		}
	}
	return gray
}

// ToGrayImage converts an image to a standard grayscale image.
func ToGrayImage(img image.Image) *image.Gray {
	bounds := img.Bounds()
	gray := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray.Set(x, y, color.GrayModel.Convert(img.At(x, y)))
		}
	}
	return gray
}
