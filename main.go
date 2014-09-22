package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"strconv"
	// Register JPEG format
	// Register PNG  format
	"log"
	"os"

	"github.com/bradfitz/iter"
	"github.com/nfnt/resize"
)

var white = color.RGBA{255, 255, 255, 0}
var black = color.RGBA{0, 0, 0, 0}
var tolerance = uint8(20)
var toResize = 10
var size = 100

type colorMap [100][100]color.Color

func isBlack(col color.RGBA) bool {
	return col.R < 10 && col.G < 10 && col.B < 10
}

func isWhite(col color.RGBA) bool {
	return col.R > 100 && col.G > 100 && col.B > 100
}

func (m *colorMap) clearNoise() *colorMap {
	for x := range iter.N(size) {
		for y := range iter.N(size) {
			m[x][y] = white
		}
	}
	return m
}

func getColorMap(i image.Image) *colorMap {
	bounds := i.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	cMap := &colorMap{}
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			cMap[x][y] = i.At(x, y)
		}
	}
	return cMap
}

func getImage(s string) image.Image {
	infile, err := os.Open(s)
	if err != nil {
		log.Fatalln(err)
	}
	defer infile.Close()

	img, err := jpeg.Decode(infile)
	if err != nil {
		log.Fatalln(err)
	}
	return img
}

func writeImage(i image.Image, filename string) *os.File {
	outfile, err := os.Create(filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer outfile.Close()
	jpeg.Encode(outfile, i, nil)
	return outfile
}

func isSimilar(diff1 color.RGBA, diff2 color.RGBA) bool {
	return diff1.R-diff2.R < tolerance && diff1.G-diff2.G < tolerance && diff1.B-diff2.B < tolerance
}

func getDiffMap(map1 *colorMap, map2 *colorMap) *colorMap {
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			diff1 := color.RGBAModel.Convert(map1[x][y]).(color.RGBA)
			diff2 := color.RGBAModel.Convert(map2[x][y]).(color.RGBA)

			if isBlack(diff1) || !isSimilar(diff1, diff2) {
				map2[x][y] = black
			}
		}
	}
	return map2
}

func getBlackPixelMap(imgMap *colorMap, oriMap *colorMap) *colorMap {
	for x := range iter.N(size) {
		for y := range iter.N(size) {
			c := color.RGBAModel.Convert(imgMap[x][y]).(color.RGBA)
			if isBlack(c) {
				oriMap[x][y] = black
			}
		}
	}
	return oriMap
}

func resizeImage(img image.Image, width int, height int) image.Image {
	return resize.Resize(uint(width), uint(height), img, resize.Lanczos3)
}

func getImageFromMap(imgMap *colorMap) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size-1, size-1))
	for x := range iter.N(size) {
		for y := range iter.N(size) {
			img.Set(x, y, imgMap[x][y])
		}
	}
	return img
}

func getBP(m *colorMap, prev *colorMap) *colorMap {
	return getBlackPixelMap(getColorMap(resizeImage(resizeImage(getImageFromMap(m), toResize, toResize), size, size)), prev)
}

func main() {
	if len(os.Args) < 4 {
		log.Fatalln("Needs min 3 arguments")
	}

	prevMap := &colorMap{}
	map1 := &colorMap{}
	t, _ := strconv.Atoi(os.Args[1])
	tolerance = uint8(t)
	for x := 2; x < len(os.Args)-1; x++ {
		if x == 2 {
			img1 := resizeImage(getImage(os.Args[x]), size, size)
			map1 = getColorMap(img1)
		} else {
			map1 = getBP(prevMap, map1)
		}
		img2 := resizeImage(getImage(os.Args[x+1]), size, size)
		map2 := getColorMap(img2)
		prevMap = getDiffMap(map1, map2)
	}
	diffImage := getImageFromMap(getBP(prevMap, map1))
	writeImage(diffImage, "images/results.jpg")

}
