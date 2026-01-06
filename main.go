package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
	"sort"
	"strings"
)

const (
	InfoColor  = "\033[92m%s\033[0m"
	ErrorColor = "\033[91m%s\033[0m"
)

func printInfo(message string) {
	fmt.Printf(InfoColor, "[INFO] ")
	fmt.Println(message)
}

func printError(message string) {
	fmt.Printf(ErrorColor, "[ERROR] ")
	fmt.Println(message)
}

func pixelSort(outputImg *image.RGBA, inputImg image.Image, luminanceThreshold uint32) {
	bounds := inputImg.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	for y := 0; y < height; y++ {
		var belowThresholdColors []color.Color
		var aboveThresholdColors []color.Color

		for x := 0; x < width; x++ {
			col := inputImg.At(x, y)
			r, g, b, _ := col.RGBA()
			intensity := 0.2126*float64(r>>8) + 0.7152*float64(g>>8) + 0.0722*float64(b>>8)

			if intensity >= float64(luminanceThreshold) {
				belowThresholdColors = append(belowThresholdColors, col)
			} else {
				aboveThresholdColors = append(aboveThresholdColors, col)
			}
		}

		sort.Slice(belowThresholdColors, func(i, j int) bool {
			r1, g1, b1, _ := belowThresholdColors[i].RGBA()
			r2, g2, b2, _ := belowThresholdColors[j].RGBA()

			intensity1 := 0.2126*float64(r1>>8) + 0.7152*float64(g1>>8) + 0.0722*float64(b1>>8)
			intensity2 := 0.2126*float64(r2>>8) + 0.7152*float64(g2>>8) + 0.0722*float64(b2>>8)

			return intensity1 < intensity2
		})

		sortedColors := append(belowThresholdColors, aboveThresholdColors...)

		for x := 0; x < width; x++ {
			outputImg.Set(x, y, sortedColors[x])
		}
	}
}

func printHelp() {
	fmt.Println("USAGE:")
	fmt.Println("\tgopxsort [FLAGS] [OPTIONS] --input <input> --output <output>\n")
	fmt.Println("FLAGS:")
	fmt.Println("\t-h, --help\t\t\tPrints help information\n")
	fmt.Println("OPTIONS:")
	fmt.Println("\t-i, --input <input>\t\tSets the input file")
	fmt.Println("\t-o, --output <output>\t\tSets the output file")
	fmt.Println("\t-t, --threshold <threshold>\tSets threshold of sorting\n")
}

func isValidThreshold(threshold uint) bool {
	return threshold >= 0 && threshold <= 255
}

func isValidFileName(fileName string) bool {
	return fileName != ""
}

func isValidImageFormat(fileName string) bool {
	validFormats := []string{".jpg", ".jpeg", ".png"}
	lowerFileName := strings.ToLower(fileName)

	for _, format := range validFormats {
		if strings.HasSuffix(lowerFileName, format) {
			return true
		}
	}

	return false
}

func main() {
	inputFileName := flag.String("input", "", "Input image file name")
	outputFileName := flag.String("output", "", "Output image file name")
	luminanceThreshold := flag.Uint("threshold", 0, "Luminance threshold (0-255)")

	flag.StringVar(inputFileName, "i", *inputFileName, "Input image file name (short)")
	flag.StringVar(outputFileName, "o", *outputFileName, "Output image file name (short)")
	flag.UintVar(luminanceThreshold, "t", *luminanceThreshold, "Luminance threshold (short)")

	helpFlag := flag.Bool("h", false, "Prints help information")
	flag.BoolVar(helpFlag, "help", *helpFlag, "Prints help information (alternative)")

	flag.Parse()

	if *helpFlag {
		printHelp()
		os.Exit(0)
	}

	if !isValidFileName(*inputFileName) || !isValidFileName(*outputFileName) {
		printError("Input and output file names are required")
		return
	}

	if !isValidThreshold(*luminanceThreshold) {
		printError("Luminance threshold must be in the range 0-255")
		return
	}

	if !isValidImageFormat(*inputFileName) {
		printError("Unsupported input image format: " + *inputFileName)
		return
	}

	if !isValidImageFormat(*outputFileName) {
		printError("Unsupported output image format: " + *outputFileName)
		return
	}

	printInfo(fmt.Sprintf("Opening file: %s", *inputFileName))
	inputFile, err := os.Open(*inputFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()

	img, _, err := image.Decode(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	sortedImg := image.NewRGBA(img.Bounds())

	printInfo(fmt.Sprintf("Sorting %dx%d image rightwards based on threshold %d", img.Bounds().Max.X, img.Bounds().Max.Y, *luminanceThreshold))
	pixelSort(sortedImg, img, uint32(*luminanceThreshold))

	printInfo(fmt.Sprintf("Saved image to: %s", *outputFileName))
	outputFile, err := os.Create(*outputFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	err = jpeg.Encode(outputFile, sortedImg, nil)
	if err != nil {
		log.Fatal(err)
	}
}
