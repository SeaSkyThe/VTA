package webcam

import (
	"bytes"
	"image"
	"image/jpeg"
	"log"
	"strings"
	"sync"

	"golang.org/x/image/draw"
)

// var charsScale = []rune{
// 	'$', '@', 'B', '%', '8', '&', 'W', 'M', '#', '*', 'o', 'a', 'h', 'k', 'b', 'd', 'p', 'q', 'w', 'm', 'Z', 'O', '0', 'Q', 'L', 'C', 'J', 'U', 'Y', 'X', 'z', 'c', 'v', 'u', 'n', 'x', 'r', 'j', 'f', 't', '/', '\\', '|', '(', ')', '1', '{', '}', '[', ']', '?', '-', '_', '+', '~', '<', '>', 'i', '!', 'l', 'I', ';', ':', ',', '"', '^', '`', '\'', '.',
// }
// var charsScale = []rune{
// 	'@', '%', '#', '*', '+', '=', '-', ':', '.', ' ',
// }

// More than 8 chars makes my art to broke :(
var charsScale = []rune{
	'@', '%', '#', '*', '+', '=', '-', ':', '.', ' ',
}

const scaleFactor = 4

// https://www.kernel.org/doc/html/latest/userspace-api/media/v4l/pixfmt-rgb.html
// We are using the format V4L2_PIX_FMT_RGB24

// Read the image, decode the []byte into JPEG

func DecodeImage(frame []byte) (image.Image, error) {
	reader := bytes.NewReader(frame)
	img, err := jpeg.Decode(reader)
	if err != nil {
		log.Printf("Error when decoding image: %s", err)
		return nil, err
	}
	return img, nil
}

func EncodeImage(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, nil)
	if err != nil {
		log.Printf("Error encoding transformed image: %s", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

func TransformToGrayScale(img image.Image) *image.Gray {
	grayImage := image.NewGray(img.Bounds())
	draw.Draw(grayImage, grayImage.Bounds(), img, image.Point{X: 0, Y: 0}, draw.Src)
	return grayImage
}

// TransformToASCII converts an image to an ASCII art string.
func TransformToASCII(img image.Image) string {
	scaledImg := ScaleImage(img, scaleFactor)
	gray := TransformToGrayScale(scaledImg)

	bounds := scaledImg.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Create a buffer to hold the ASCII art.
	var buffer strings.Builder
	buffer.Grow((width + 1) * height) // preallocate memory

	// Channel for each row's ASCII data
	type row struct {
		y     int
		ascii string
	}
	rowsChan := make(chan row, height)

	// Process rows in parallel
	var wg sync.WaitGroup
	for y := 0; y < height; y++ {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()
			var line strings.Builder
			for x := 0; x < width; x = x + 1 {
				grayValue := gray.GrayAt(x, y).Y
				line.WriteByte(byte(grayScaleToChar(uint8(grayValue))))
			}
			rowsChan <- row{y, line.String()}
		}(y)
	}

	// Close the channel once all rows are processed
	go func() {
		wg.Wait()
		close(rowsChan)
	}()

	// Collect rows in order
	asciiRows := make([]string, height)
	for r := range rowsChan {
		asciiRows[r.y] = r.ascii
	}

	// Join all rows into the final ASCII art
	for _, row := range asciiRows {
		buffer.WriteString(row)
		buffer.WriteByte('\n')
	}

	return buffer.String()
}

func ScaleImage(src image.Image, scaleFactor int) image.Image {
	srcBounds := src.Bounds()
	dstBounds := image.Rect(0, 0, srcBounds.Dx()/scaleFactor, srcBounds.Dy()/scaleFactor)
	dst := image.NewGray(dstBounds)
	draw.BiLinear.Scale(dst, dstBounds, src, srcBounds, draw.Src, nil)
	return dst
}

// Function to map a greyscale value to a character
func grayScaleToChar(grey uint8) rune {
	numChars := len(charsScale)
    // Scaling factor based on the number of chars
    scalingFactor := 256.0 / float64(numChars)
	// Calculate the index by scaling the grayscale value to the number of characters
	index := int(float64(grey)/scalingFactor)

	// Ensure index stays within bounds
	if index < 0 {
		index = 0
	} else if index >= numChars {
		index = numChars - 1
	}

	return charsScale[index]
}
