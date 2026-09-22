package backend

import (
	"fmt"
	"image"
	"os"

	_ "image/jpeg"
	_ "image/png"
)

type ChromakeyService struct{}

// CornerColorFromFile：提供 Wails / Svelte 呼叫。
// inputPath 必須是本機檔案系統的絕對路徑。
func (service *ChromakeyService) CornerColorFromFile(inputPath string, sampleSize int) (ColorResult, error) {

	file, err := os.Open(inputPath)
	if err != nil {
		return ColorResult{}, fmt.Errorf("open input image: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return ColorResult{}, fmt.Errorf("decode input image: %w", err)
	}

	r, g, b := sampleCornerColor(img, sampleSize)

	return ColorResult{
		R:   r,
		G:   g,
		B:   b,
		Hex: fmt.Sprintf("#%02X%02X%02X", r, g, b),
	}, nil
}

// ProcessFile：提供 Wails / Svelte 呼叫。
// p 需要是公開、可 JSON 序列化的 KeyParams。
func (service *ChromakeyService) ProcessFile(inputPath string, outputPath string, p KeyParams) error {
	return processFile(inputPath, outputPath, p)
}
