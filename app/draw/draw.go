package draw

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
)

var (
	bgPath, fontPath, overlayPath, outputPath string
)

func Render(imgPath, text string) string {
	loadDirectories()
	background := openImage(bgPath)
	if background == nil {
		return ""
	}

	dc := gg.NewContextForImage(background)
	drawText(dc, text, fontPath)
	destImg := drawImg(dc, imgPath)
	destImg = drawOverlay(destImg, overlayPath)

	return saveImage(outputPath, destImg)
}

func loadDirectories() {
	rootDir, err := os.Getwd()
	if checkError("Getting working directory", err) {
		return
	}

	outputPath = filepath.Join(rootDir, "temp", "output.png")
	rootDir = filepath.Dir(rootDir)
	bgPath = filepath.Join(rootDir, "resources", "BG.png")
	fontPath = filepath.Join(rootDir, "resources", "Buran USSR.ttf")
	overlayPath = filepath.Join(rootDir, "resources", "Overlay.png")
}

func openImage(path string) image.Image {
	file, err := os.Open(path)
	if checkError("Opening file", err) {
		return nil
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if checkError("Decoding image", err) {
		return nil
	}
	return img
}

func saveImage(path string, img image.Image) string {
	file, err := os.Create(path)
	if checkError("Creating file", err) {
		return ""
	}
	defer file.Close()

	err = png.Encode(file, img)
	if checkError("Saving image", err) {
		return ""
	}
	return path
}

func drawImg(context *gg.Context, path string) image.Image {
	img := openImage(path)
	if img == nil {
		return context.Image()
	}

	img = cropToSquare(img)
	targetSize := context.Width() * 6 / 10
	img = resize.Resize(uint(targetSize), uint(targetSize), img, resize.Lanczos3)

	centerX := (context.Width() - img.Bounds().Dx()) / 2
	centerY := (context.Height() - img.Bounds().Dy()) / 2
	destImg := image.NewRGBA(context.Image().Bounds())
	draw.Draw(destImg, destImg.Bounds(), context.Image(), image.Point{}, draw.Over)
	draw.Draw(destImg, img.Bounds().Add(image.Pt(centerX, centerY)), img, image.Point{}, draw.Over)

	return destImg
}

func drawOverlay(baseImg image.Image, path string) image.Image {
	overlay := openImage(path)
	if overlay == nil {
		return baseImg
	}

	baseRGBA := image.NewRGBA(baseImg.Bounds())
	draw.Draw(baseRGBA, baseRGBA.Bounds(), baseImg, image.Point{}, draw.Over)

	overlayRGBA := image.NewRGBA(overlay.Bounds())
	for y := 0; y < overlay.Bounds().Dy(); y++ {
		for x := 0; x < overlay.Bounds().Dx(); x++ {
			originalColor := overlay.At(x, y)
			r, g, b, a := originalColor.RGBA()
			alpha := uint16(float64(a) * 0.6)

			overlayRGBA.Set(x, y, color.NRGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(alpha >> 8),
			})
		}
	}

	centerX := (baseRGBA.Bounds().Dx() - overlay.Bounds().Dx()) / 2
	centerY := (baseRGBA.Bounds().Dy() - overlay.Bounds().Dy()) / 2
	draw.Draw(baseRGBA, overlay.Bounds().Add(image.Pt(centerX, centerY)), overlayRGBA, image.Point{}, draw.Over)

	return baseRGBA
}

func drawText(context *gg.Context, text, fontPath string) {
	if err := context.LoadFontFace(fontPath, fontSize(context.Width(), context.Height())); checkError("Loading font", err) {
		return
	}

	context.SetColor(color.RGBA{33, 35, 50, 255})
	context.DrawStringAnchored(text, float64(context.Width())/2, float64(context.Height())*0.86, 0.5, 0.5)
}

func fontSize(width, height int) float64 {
	return float64(max(width, height) / 1000 * 85)
}

func cropToSquare(img image.Image) image.Image {
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	if width == height {
		return img
	}

	cropRect := image.Rect(0, 0, width, height)
	if width > height {
		cropRect = image.Rect((width-height)/2, 0, (width+height)/2, height)
	} else {
		cropRect = image.Rect(0, (height-width)/2, width, (height+width)/2)
	}

	rgbaImg := image.NewRGBA(img.Bounds())
	draw.Draw(rgbaImg, rgbaImg.Bounds(), img, image.Point{}, draw.Over)
	return rgbaImg.SubImage(cropRect).(*image.RGBA)
}

func checkError(msg string, err error) bool {
	if err != nil {
		fmt.Println(msg+":", err)
		return true
	}
	return false
}

func ResizeImage(inputPath string, maxSize uint) (string, error) {
	file, err := os.Open(inputPath)
	if checkError("Opening file", err) {
		return "", err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if checkError("Decoding image", err) {
		return "", err
	}

	resizedImg := resize.Resize(maxSize, 0, img, resize.Lanczos3)
	outputPath := inputPath + "_resized.png"
	outFile, err := os.Create(outputPath)
	if checkError("Creating file", err) {
		return "", err
	}
	defer outFile.Close()

	err = png.Encode(outFile, resizedImg)
	if checkError("Encoding image", err) {
		return "", err
	}

	return outputPath, nil
}
