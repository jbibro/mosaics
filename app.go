package main

import (
	"flag"
	"github.com/nfnt/resize"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
)

func main() {
	thumbnailsDir := flag.String("thumbnailsDir", ".", "Thumbnails Directory")
	thumbnailEdgeSize := flag.Uint("thumbnailEdgeSize", 30, "Thumbnail Edge Size (px)")
	input := flag.String("input", "", "Input file")
	flag.Parse()

	thumbnails := thumbnails(*thumbnailsDir)
	thumbnailColors := make(map[image.Image]color.Color)
	for _, t := range thumbnails {
		file, _ := os.Open(*thumbnailsDir + "/" + t.Name())
		src, err := jpeg.Decode(file)
		if err == nil {
			avgColor := averageColor(src, src.Bounds())
			thumbnailColors[src] = avgColor
		}
	}

	file, e := os.Open(*input)
	if e != nil {
		log.Fatal("unable to open file")
	}
	src, e := jpeg.Decode(file)
	if e != nil {
		log.Fatal("unable to decode file to jpg image")
	}

	dest := image.NewRGBA(image.Rect(0, 0, src.Bounds().Max.X, src.Bounds().Max.Y))

	for j := 0; j < src.Bounds().Max.Y; j += int(*thumbnailEdgeSize) {
		for i := 0; i < src.Bounds().Max.X; i += int(*thumbnailEdgeSize) {
			srcSubimage := image.Rect(i, j, i+int(*thumbnailEdgeSize), j+int(*thumbnailEdgeSize))
			imageColor := averageColor(src, srcSubimage)
			bestImage := bestMatchingImage(imageColor, thumbnailColors)
			resized := resize.Resize(*thumbnailEdgeSize, *thumbnailEdgeSize, bestImage, resize.NearestNeighbor)
			startingPoint := image.Point{i, j}
			r := image.Rectangle{startingPoint, startingPoint.Add(srcSubimage.Size())}
			draw.Draw(dest, r, resized, image.Point{0, 0}, draw.Src)
		}
	}

	created, e := os.Create("mosaic.jpg")
	if e != nil {
		log.Fatal("unable to create file")
	}
	_ = jpeg.Encode(created, dest, &jpeg.Options{Quality: 100})
}

func averageColor(image image.Image, rectangle image.Rectangle) color.Color {
	var avgr, avgg, avgb uint32
	size := uint32(30)

	for i := 0; i < int(size); i++ {
		randX := rand.Intn(rectangle.Max.X-rectangle.Min.X) + rectangle.Min.X
		randY := rand.Intn(rectangle.Max.Y-rectangle.Min.Y) + rectangle.Min.Y

		r, g, b, _ := image.At(randX, randY).RGBA()
		avgr += r
		avgg += g
		avgb += b
	}

	avgr /= size
	avgg /= size
	avgb /= size
	return color.RGBA{uint8(avgr / 0x101), uint8(avgg / 0x101), uint8(avgb / 0x101), 255}

}

func thumbnails(dir string) []os.FileInfo {
	files, _ := ioutil.ReadDir(dir)
	return files
}

func bestMatchingImage(targetColor color.Color, thumbnails map[image.Image]color.Color) image.Image {
	var ret image.Image
	min := math.MaxFloat64
	for i, c := range thumbnails {
		d := distance(targetColor, c)
		if d < min {
			min = d
			ret = i
		}
	}
	return ret
}

func distance(c1, c2 color.Color) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	return math.Sqrt(math.Pow(float64(r2-r1), 2) + math.Pow(float64(g2-g1), 2) + math.Pow(float64(b2-b1), 2))
}
