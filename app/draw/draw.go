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

var bgPath, fontPath, overlayPath, outputPath string

func main() {
	loadDirectories()
	const (
		imgPath = "C:\\Users\\user\\Dropbox\\Programming\\Go\\PostInator\\resources\\test.jpg"
		text    = "main.go"
	)

	background := openImage(bgPath)
	if background == nil {
		return
	}

	dc := gg.NewContextForImage(background)
	drawText(dc, text, fontPath)
	destImg := drawImg(dc, imgPath)
	dc = gg.NewContextForImage(destImg)
	destImg = drawOverlay(dc, overlayPath)

	saveImage(outputPath, destImg)
}

func Render(imgPath, text string) string {
	loadDirectories()
	background := openImage(bgPath)
	if background == nil {
		fmt.Println("Ошибка загрузки фона")
		return ""
	}

	dc := gg.NewContextForImage(background)
	drawText(dc, text, fontPath)
	destImg := drawImg(dc, imgPath)
	dc = gg.NewContextForImage(destImg)
	destImg = drawOverlay(dc, overlayPath)

	return saveImage(outputPath, destImg)

}

func loadDirectories() {
	rootDir, _ := os.Getwd()
	outputPath = filepath.Join(rootDir, "temp", "output.png")
	rootDir = filepath.Dir(rootDir)
	bgPath = filepath.Join(rootDir, "resources", "BG.png")
	fontPath = filepath.Join(rootDir, "resources", "Buran USSR.ttf")
	overlayPath = filepath.Join(rootDir, "resources", "Overlay.png")
}

func openImage(path string) image.Image {
	file, err := os.Open(path)
	if checkError("Open file:", err) {
		return nil
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if checkError("Decoding:", err) {
		return nil
	}
	return img
}

func saveImage(path string, img image.Image) string {
	file, err := os.Create(path)
	checkError("Creating file:", err)

	defer file.Close()

	err = png.Encode(file, img)
	checkError("Saving image:", err)

	return path
}

func drawImg(context *gg.Context, path string) image.Image {
	width, height := context.Width(), context.Height()
	img := openImage(path)
	if img == nil {
		return context.Image()
	}

	img = cropToSquare(img)
	targetSize := width * 6 / 10
	img = resize.Resize(uint(targetSize), uint(targetSize), img, resize.Lanczos3)

	centerX := (width - img.Bounds().Dx()) / 2
	centerY := (height - img.Bounds().Dy()) / 2

	destImg := image.NewRGBA(context.Image().Bounds())
	draw.Draw(destImg, destImg.Bounds(), context.Image(), image.Point{}, draw.Over)
	offset := image.Pt(centerX, centerY)
	draw.Draw(destImg, img.Bounds().Add(offset), img, image.Point{}, draw.Over)

	return destImg
}

func drawOverlay(context *gg.Context, path string) image.Image {
	baseImg := context.Image()
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
	offset := image.Pt(centerX, centerY)

	draw.Draw(baseRGBA, overlay.Bounds().Add(offset), overlayRGBA, image.Point{}, draw.Over)

	return baseRGBA
}

func drawText(context *gg.Context, text, fontPath string) {
	width, height := context.Width(), context.Height()

	if err := context.LoadFontFace(fontPath, fontSize(width, height)); checkError("Loading font:", err) {
		return
	}

	context.SetColor(color.RGBA{33, 35, 50, 255})
	context.DrawStringAnchored(text, float64(width)/2, float64(height)*0.86, 0.5, 0.5)
}

func fontSize(width, height int) float64 {
	return float64(max(width, height) / 1000 * 85)
}

func cropToSquare(img image.Image) image.Image {
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	if width == height {
		return img
	}

	var cropRect image.Rectangle
	if width > height {
		cropX := (width - height) / 2
		cropRect = image.Rect(cropX, 0, cropX+height, height)
	} else {
		cropY := (height - width) / 2
		cropRect = image.Rect(0, cropY, width, cropY+width)
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
	if err != nil {
		return "", err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}

	resizedImg := resize.Resize(maxSize, 0, img, resize.Lanczos3)

	outputPath := inputPath + "_resized.png"
	outFile, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	err = png.Encode(outFile, resizedImg)
	if err != nil {
		return "", err
	}

	return outputPath, nil
}
