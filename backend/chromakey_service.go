package backend

import (
	"fmt"
	"image"
	"os"

	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

// 是提供給 Wails 前端呼叫的 Chroma Key 服務
type ChromakeyService struct{}

// 讀取指定圖片，從四個角落取樣，回傳平均背景色
//
// 參數：
//   - inputPath：來源圖片的本機檔案路徑
//   - sampleSize：每個角落取樣區塊的邊長，單位為像素
//
// 回傳：
//   - ColorResult：包含平均背景色的 R、G、B 與 Hex 欄位
//   - error：開檔或圖片解碼失敗時回傳錯誤
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
	hex := fmt.Sprintf("#%02X%02X%02X", r, g, b)

	return ColorResult{R: r, G: g, B: b, Hex: hex}, nil
}

// 對來源圖片執行 Chroma Key 去背，並輸出透明背景 PNG
//
// 參數：
//   - inputPath：來源圖片的本機檔案路徑
//   - outputPath：處理完成後的透明 PNG 輸出路徑
//   - p：Chroma Key 的去背參數
//
// 回傳：
//   - nil：去背成功，PNG 已寫入 outputPath
//   - error：若來源檔不存在、無法解碼、輸出路徑無法建立或 PNG 寫入失敗
func (service *ChromakeyService) ProcessFile(inputPath string, outputPath string, p KeyParams) error {
	return processFile(inputPath, outputPath, p)
}
