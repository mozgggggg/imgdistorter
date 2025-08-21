package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color/palette"
	"image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"

	"image/draw"
	pxllist "imgdistorter/pixellist"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

const (
	miscolorintense = 0.11
	miscolorspread  = 10
	//miscolor - changes the colors of [miscolorspread]% pixels with a [miscolorintense]% intensivity, better not be over like 3 or 5, basically a background noise

	chunksize          = 0.2
	chunkcount         = 7
	verticalswaponly   = false
	horizontalswaponly = true
	//swapchunk - swaps chunks of pixels with about [chunksize]% of image in each chunk with the other random chunk of image, please dont do chunksize more than 1 it will likely break, vertical/horizontal only swaps do as they say - if true only swaps chunck in the mentioned direction

	horizontalrows = false
	verticalrows   = true
	rowcount       = 50
	//making rows/columns with width of 1 pixel entirely one of 8: red yellow green blue cyan violet black or white, if horizontal/vertical rows is false - doesnt draw these

	spillradius    = 5
	spillfrequency = 0.1
	infectchance   = 0.7

	censorsize  = 0.2
	censorcount = 7

	applycolorred   = -0.1
	applycolorgreen = 0.3
	applycolorblue  = -0.1
)

func miscolorinit(matrix [][][]int, width int, height int) {
	for i := 0; i < int(float64(height*width)*miscolorspread); i++ {
		wrand := rand.Intn(width)
		hrand := rand.Intn(height)
		//getting pixel

		matrix[hrand][wrand][0] += int(float64(rand.Intn(65535)-32767) * miscolorintense)
		matrix[hrand][wrand][1] += int(float64(rand.Intn(65535)-32767) * miscolorintense)
		matrix[hrand][wrand][2] += int(float64(rand.Intn(65535)-32767) * miscolorintense)
		//changing color

		if matrix[hrand][wrand][0] < 0 {
			matrix[hrand][wrand][0] = 0
		}
		if matrix[hrand][wrand][1] < 0 {
			matrix[hrand][wrand][1] = 0
		}
		if matrix[hrand][wrand][2] < 0 {
			matrix[hrand][wrand][2] = 0
		}
		if matrix[hrand][wrand][0] > 65535 {
			matrix[hrand][wrand][0] = 65535
		}
		if matrix[hrand][wrand][1] > 65535 {
			matrix[hrand][wrand][1] = 65535
		}
		if matrix[hrand][wrand][2] > 65535 {
			matrix[hrand][wrand][2] = 65535
		}
		// if the color is off limits, make it actually acceptable on a 16-bit scale
	}
}

func swapchuckinit(matrix [][][]int, width int, height int) {
	for i := 0; i < chunkcount; i++ {
		chunkheight := rand.Intn(height)
		chunkwidth := int((1 - float64(chunkheight/height)) * float64(width))
		chunkheight = int(float64(chunkheight) * chunksize)
		chunkwidth = int(float64(chunkwidth) * chunksize)
		//making and calculating approximate chunk size

		chunkx := rand.Intn(width - chunkwidth)
		chunky := rand.Intn(height - chunkheight)
		//the 1st chunk left top corner defining
		chunkx2 := chunkx
		if !verticalswaponly {
			chunkx2 = rand.Intn(width - chunkwidth)
		}

		chunky2 := chunky
		if !horizontalswaponly {
			chunky2 = rand.Intn(height - chunkheight)
		}
		//defining the 2nd swap chunk + checking the directions of the swap
		for i := 0; i < chunkheight; i++ {
			for l := 0; l < chunkwidth; l++ {
				matrix[chunky+i][chunkx+l], matrix[chunky2+i][chunkx2+l] = matrix[chunky2+i][chunkx2+l], matrix[chunky+i][chunkx+l]
			}
		}
		//the swapping itself
	}
}

func deadrowsinit(matrix [][][]int, width int, height int) {
	for i := 0; i < rowcount; i++ {
		direction := true
		if !verticalrows {
			direction = false
		}

		if horizontalrows && verticalrows {
			direction = rand.Intn(2) == 1
		}
		red := rand.Intn(2) * 65535
		green := rand.Intn(2) * 65535
		blue := rand.Intn(2) * 65535

		if direction {
			row := rand.Intn(width)
			for i := 0; i < height; i++ {
				matrix[i][row][0] = red
				matrix[i][row][1] = green
				matrix[i][row][2] = blue
			}

		} else {
			row := rand.Intn(height)
			for i := 0; i < width; i++ {
				matrix[row][i][0] = red
				matrix[row][i][1] = green
				matrix[row][i][2] = blue
			}
		}
	}
}

func liquifyinit(matrix [][][]int, width int, height int) {
	for i := 0; i < int(spillfrequency*float64(width)*float64(height)); i++ {
		if width > spillradius && height > spillradius {
			spillx := rand.Intn(width-((spillradius+1)*2)) + spillradius + 1
			spilly := rand.Intn(height-((spillradius+1)*2)) + spillradius + 1
			for x := (spillradius * -1); x < spillradius; x++ {
				for y := (spillradius * -1); y < spillradius; y++ {
					if rand.Intn(100) < int(infectchance*100) {
						matrix[spilly+y][spillx+x][0] = matrix[spilly][spillx][0]
						matrix[spilly+y][spillx+x][1] = matrix[spilly][spillx][1]
						matrix[spilly+y][spillx+x][2] = matrix[spilly][spillx][2]
					}
				}
			}

		}
	}
}

func censorinit(matrix [][][]int, width int, height int) {
	for i := 0; i < chunkcount; i++ {
		chunkheight := rand.Intn(height)
		chunkwidth := int((1 - float64(chunkheight/height)) * float64(width))
		chunkheight = int(float64(chunkheight) * chunksize)
		chunkwidth = int(float64(chunkwidth) * chunksize)
		//making and calculating approximate chunk size

		chunkx := rand.Intn(width - chunkwidth)
		chunky := rand.Intn(height - chunkheight)

		for i := 0; i < chunkheight; i++ {
			for l := 0; l < chunkwidth; l++ {
				matrix[chunky+i][chunkx+l][0], matrix[chunky+i][chunkx+l][1], matrix[chunky+i][chunkx+l][2] = 0, 0, 0

			}
		}

	}
}

func applycolorinit(matrix [][][]int, width int, height int) {
	for i := 0; i < height; i++ {
		for l := 0; l < width; l++ {
			matrix[i][l][0] += int(math.Floor(65535 * applycolorred))
			matrix[i][l][1] += int(math.Floor(65535 * applycolorgreen))
			matrix[i][l][2] += int(math.Floor(65535 * applycolorblue))
			if matrix[i][l][0] < 0 {
				matrix[i][l][0] = 0
			}
			if matrix[i][l][1] < 0 {
				matrix[i][l][1] = 0
			}
			if matrix[i][l][2] < 0 {
				matrix[i][l][2] = 0
			}
			if matrix[i][l][0] > 65535 {
				matrix[i][l][0] = 65535
			}
			if matrix[i][l][1] > 65535 {
				matrix[i][l][1] = 65535
			}
			if matrix[i][l][2] > 65535 {
				matrix[i][l][2] = 65535
			}
		}
	}

}

func main() {
	fmt.Println("To Start, Type [start [sequence]] to modify an your selected image/GIF (!!!RENAME TO image.png/.jpg/.jpeg/.gif)")
	fmt.Println("where [sequence] is a sequence of letters where each letter represents a filter (the filters are applied in order)")
	fmt.Println("M - Miscolor")
	fmt.Println("S - Swapchunk")
	fmt.Println("D - Deadrows")
	fmt.Println("L - Liquify")
	fmt.Println("C - Censor")
	fmt.Println("Examples: start MLDC, modify SDLC")

	//scanning user input
	var userinputraw string

	scanner := bufio.NewScanner(os.Stdin)

	// prompt the user for input
	fmt.Print("Enter input: ")

	if scanner.Scan() {
		userinputraw = scanner.Text()
	}

	userinput := strings.Fields(strings.ToLower(userinputraw))

	var file *os.File
	var err error

	// opening file
	format := ""
	imageExtensions := []string{".jpg", ".png", ".jpeg", ".gif"}

	for _, num := range imageExtensions {
		file, err = os.Open("image" + num)
		if err != nil && !os.IsNotExist(err) {
			panic(err)
		}
		if err == nil {
			format = num
			defer file.Close()
			break
		}
	}
	if userinput[0] == "start" {

		// decoding image
		fmt.Println(format)
		if format != ".gif" {
			img, _, err := image.Decode(file)
			if err != nil {
				panic(err)
			}
			// getting matrix
			matrix := pxllist.Getlist(img)

			bounds := img.Bounds()
			width := bounds.Dx()
			height := bounds.Dy()

			for i := 0; i < len(userinput[1]); i++ {
				switch string(userinput[1][i]) {
				case "m":
					miscolorinit(matrix, width, height)
				case "s":
					swapchuckinit(matrix, width, height)
				case "d":
					deadrowsinit(matrix, width, height)
				case "l":
					liquifyinit(matrix, width, height)
				case "c":
					censorinit(matrix, width, height)
				case "a":
					applycolorinit(matrix, width, height)
				}

			}

			file, err := os.Create("result.png")
			if err != nil {
				panic(err)
			}
			defer file.Close()
			if err := png.Encode(file, pxllist.Savelist(matrix)); err != nil {
				panic(err)
			}

			fmt.Println("Successfully Edited Image!")

		} else {
			//if the thing were modifying is actually a .gif
			g, err := gif.DecodeAll(file)
			if err != nil {
				fmt.Println("Error decoding GIF:", err)
				return
			}

			// getting matrix

			frames := []image.Image{}

			var gifImages []*image.Paletted
			var delays []int

			for _, img := range g.Image {
				matrix := pxllist.Getlist(img)

				bounds := img.Bounds()
				width := bounds.Dx()
				height := bounds.Dy()
				if len(userinput) > 1 {
					for i := 0; i < len(userinput[1]); i++ {
						switch string(userinput[1][i]) {
						case "m":
							miscolorinit(matrix, width, height)
						case "s":
							swapchuckinit(matrix, width, height)
						case "d":
							deadrowsinit(matrix, width, height)
						case "l":
							liquifyinit(matrix, width, height)
						case "c":
							censorinit(matrix, width, height)
						case "a":
							applycolorinit(matrix, width, height)
						}
					}
				}

				frames = append(frames, pxllist.Savelist(matrix))
			}
			for i, img := range frames {
				palettedImg := image.NewPaletted(img.Bounds(), palette.Plan9)
				draw.FloydSteinberg.Draw(palettedImg, img.Bounds(), img, image.Point{})

				gifImages = append(gifImages, palettedImg)
				delays = append(delays, g.Delay[i])
			}
			fmt.Println("Successfully Edited Image!")
			outFile, err := os.Create("result.gif")
			if err != nil {
				panic(err)
			}
			defer outFile.Close()

			// Encode the GIF
			err = gif.EncodeAll(outFile, &gif.GIF{
				Image: gifImages,
				Delay: delays,
			})
			if err != nil {
				panic(err)
			}
		}
	}
	if userinput[0] == "function" {
		if len(userinput) > 1 {
			if userinput[1] == "atpixel" {
				var userinputraw string

				scannercoords := bufio.NewScanner(os.Stdin)
				fmt.Print("Enter coords: ")

				if scannercoords.Scan() {
					userinputraw = scannercoords.Text()
				}

				userinput := strings.Fields(userinputraw)
				if format != ".gif" {
					if len(userinput) > 1 {
						img, _, err := image.Decode(file)
						if err != nil {
							panic(err)
						}
						// getting matrix
						matrix := pxllist.Getlist(img)

						bounds := img.Bounds()
						width := bounds.Dx()
						height := bounds.Dy()
						xcheck, err := strconv.Atoi(userinput[0])
						ycheck, err := strconv.Atoi(userinput[1])

						if xcheck <= width && ycheck <= height {
							fmt.Println(matrix[ycheck][xcheck])
						}
					}
				} else {
					fmt.Println("doesnt work on gifs")
				}
			}
		} else {
			fmt.Println("Please name the function too")
		}
	}
}
