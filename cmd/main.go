package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"os"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type Pixel struct {
	R, G, B uint8 //intensity of each color is store as 8 bit value
}

// this function return the pixel matrix and each pixel with rgb values
func getPixelMatrix(img image.Image) [][]Pixel {
	// var pixelMatrix = [][]int
	//get the image boundaries
	// fmt.Println("image bound", img.Bounds())

	bounds := img.Bounds() // (0,0)-(182,240)
	//extracting the width and height of the image in pixels
	width, height := bounds.Max.X, bounds.Max.Y
	// fmt.Println("height", height)
	// fmt.Println("width", width)
	// get the pixel matrix

	pixelMatrix := make([][]Pixel, height)
	for y := 0; y < height; y++ {
		pixelMatrix[y] = make([]Pixel, width)
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA() // retrieves the color of the  pixel or at that coordinate
			pixelMatrix[y][x] = Pixel{
				// The image.Image interface in Go provides pixel color values as 16-bit integers (type uint32), ranging from 0 to 65535.
				// Most image formats (like JPEG) use 8-bit channels.
				// To work with such images, we need to convert the 16-bit values (range 0–65535) into 8-bit values (range 0–255).
				// we can do this by performing right shift . In go >> is used as right shift operator
				// wrapping the shifting result inside uint8 () is to ensure value is of type uint8

				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
			}
		}
	}

	return pixelMatrix
}

// function to convert the rgb value of each pixel into grayscale
func rgbToGrayScale(pixelMatrix [][]Pixel) [][]uint8 {
	height := len(pixelMatrix)
	width := len(pixelMatrix[0])
	grayMatrix := make([][]uint8, height)

	for i := range grayMatrix {
		grayMatrix[i] = make([]uint8, width)
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := pixelMatrix[y][x]
			// converting into gray scale using luminosity  method, also could use lightness method or others .
			// We’re more sensitive to green than other colors, so green is weighted most heavily in luminosity method
			gray := 0.21*float64(pixel.R) + 0.72*float64(pixel.G) + 0.07*float64(pixel.B)
			grayMatrix[y][x] = uint8(gray)
		}
	}
	return grayMatrix
}

func saveGrayScaleImage(grayMatrix [][]uint8) {
	height := len(grayMatrix)
	width := len(grayMatrix[0])

	// creating new grayscale image
	img := image.NewGray(image.Rect(0, 0, width, height))

	// filling the grayscale image with pixel data
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			gray := grayMatrix[y][x]
			img.Set(x, y, color.Gray{Y: gray})
		}
	}

	//saving the grayscale image to a file
	file, err := os.Create("outputs/grayImages/output1.png")

	if err != nil {
		fmt.Println("unable to save grayscale image", err)
		return
	}
	defer file.Close()

	//encoding the image as png image
	err = png.Encode(file, img)
	if err != nil {
		fmt.Println("unable to encode grayscale image to png format")
		return
	}

}

func grayscaleAsciiImage(grayMatrix [][]uint8) []string {
	asciiChars := ".:-=+*#%@"
	height := len(grayMatrix)
	width := len(grayMatrix[0])

	aspectRatio := 2.0
	sampleWidth := width
	sampleHeight := int(float64(height) / aspectRatio)

	stepX := float64(width) / float64(sampleWidth)
	stepY := float64(height) / float64(sampleHeight)

	asciiLines := make([]string, sampleHeight)
	for y := 0; y < sampleHeight; y++ {
		var lineBuilder strings.Builder
		for x := 0; x < sampleWidth; x++ {
			sampleY := int(float64(y) * stepY)
			sampleX := int(float64(x) * stepX)

			if sampleY >= height {
				sampleY = height - 1
			}
			if sampleX >= width {
				sampleX = width - 1
			}
			gray := grayMatrix[sampleY][sampleX]
			// Map grayscale value (0-255) to an index in the asciiChars string
			index := int(float64(gray) / 255.0 * float64(len(asciiChars)-1))
			// fmt.Print(string(asciiChars[index]))
			lineBuilder.WriteString(string(asciiChars[index]))

		}
		asciiLines[y] = lineBuilder.String()

		// fmt.Println()
	}
	return asciiLines

}

func saveAsciiArtAsImage(asciiLines []string, outputPath string) error {
	//  font parameters
	face := basicfont.Face7x13
	fontWidth := 7   // width of each character
	fontHeight := 13 // height of each character
	padding := 10    // padding around the text

	// Calculating the maximum line length
	maxLineLength := 0
	for _, line := range asciiLines {
		if len(line) > maxLineLength {
			maxLineLength = len(line)
		}
	}

	// Calculating image dimensions with padding
	imgWidth := fontWidth*maxLineLength + 2*padding
	imgHeight := fontHeight*len(asciiLines) + 2*padding

	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	// white backgound  banauna
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Transparent}, image.Point{}, draw.Src)

	// Creating a drawer for text
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.White), // Text color
		Face: face,
	}

	// Drawing the ASCII art text
	for i, line := range asciiLines {
		// Set position for drawing text
		drawer.Dot = fixed.Point26_6{
			X: fixed.I(padding),
			Y: fixed.I(padding + (i+1)*fontHeight),
		}

		drawer.DrawString(line)
	}

	// Creating the output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		return fmt.Errorf("failed to encode image: %v", err)
	}

	return nil
}

func main() {
	//load the image
	file, err := os.Open("assets/picka.png")
	if err != nil {
		fmt.Println("error is", err)
	}
	defer file.Close()

	// decoding the image
	// second parameter give the file type in string format
	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Println("error while decoding image", err)
	}
	// fmt.Println("img is", img)
	//get the pixel matrix of the image
	pixelMatrix := getPixelMatrix(img)

	// converting into grayscale
	grayMatrix := rgbToGrayScale(pixelMatrix)

	//saving the grayscale image
	saveGrayScaleImage(grayMatrix)

	// ascii image

	asciiArt := grayscaleAsciiImage(grayMatrix)

	imageFilePath := "outputs/asciiImages/output1.png"
	err = saveAsciiArtAsImage(asciiArt, imageFilePath)
	if err != nil {
		fmt.Println("Unable to save ASCII art as image:", err)
	} else {
		fmt.Println("ASCII art saved as image:", imageFilePath)
	}
}
