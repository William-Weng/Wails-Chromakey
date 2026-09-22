package backend

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"

	_ "image/jpeg"
)

// 讀取來源圖片檔案，套用 Chroma Key 去背，並輸出透明背景 PNG
//
// 參數：
//
//   - inputPath：來源圖片的本機絕對或相對檔案路徑
//   - outputPath：輸出 PNG 的目標檔案路徑
//   - params：Chroma Key 去背設定。
//
// 回傳：
//
//   - nil：圖片處理與 PNG 輸出成功
//   - error：開檔、建立輸出檔、圖片解碼、去背或 PNG 編碼失敗
func sampleCornerColor(img image.Image, sampleSize int) (r, g, b uint8) {

	bounds := img.Bounds()
	sampleSize, err := safeSampleSize(bounds, sampleSize)

	if err != nil {
		return 0, 0, 0
	}

	corners := cornersCoordinate(bounds, sampleSize)

	return averageColor(img, sampleSize, corners)
}

// 從圖片四個角落取樣，計算並回傳平均 RGB 顏色
//
// 參數：
//   - img：已解碼的來源圖片
//   - sampleSize：每一個角落取樣區域的正方形邊長，單位為像素
//
// 回傳：
//   - r：四個角落所有有效取樣像素的平均紅色通道值，範圍 0～255
//   - g：四個角落所有有效取樣像素的平均綠色通道值，範圍 0～255
//   - b：四個角落所有有效取樣像素的平均藍色通道值，範圍 0～255
func processFile(inputPath string, outputPath string, params KeyParams) error {

	input, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open input image: %w", err)
	}
	defer input.Close()

	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output image: %w", err)
	}
	defer output.Close()

	return processReaderWriter(input, output, params)
}

// 對來源圖片套用 Chroma Key 去背，回傳含透明度的 NRGBA 圖片
//
// 演算法會逐像素比較原圖顏色與指定的 Key Color：
//   - 與 Key Color 足夠接近的像素：視為背景，alpha 設為 0（完全透明）
//   - 與 Key Color 差異夠大的像素：視為前景，alpha 設為 255（完全不透明）
//   - 若設定 Softness：介於背景與前景之間的像素，會產生漸層 alpha，讓邊緣較柔和，減少硬邊與鋸齒感
//
// params 包含：
//   - KeyR / KeyG / KeyB：欲移除背景的 RGB 顏色
//   - Tolerance：與 Key Color 的基本顏色距離閾值
//   - Softness：從透明到不透明的額外柔化距離
//
// 輸出使用 *image.NRGBA，因為它直接儲存 R、G、B、A 四個 8-bit channel，適合後續以 png.Encode() 輸出為帶透明背景的 PNG
func keyImage(src image.Image, params KeyParams) *image.NRGBA {

	bounds := src.Bounds()
	out := image.NewNRGBA(bounds)

	keyR := float64(params.KeyR)
	keyG := float64(params.KeyG)
	keyB := float64(params.KeyB)

	toleranceSquared := params.Tolerance * params.Tolerance

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {

			pixel := src.At(x, y)
			r16, g16, b16, _ := pixel.RGBA()

			r8 := uint8(r16 >> 8)
			g8 := uint8(g16 >> 8)
			b8 := uint8(b16 >> 8)

			alpha := keyAlpha(params, toleranceSquared, keyR, keyG, keyB, r8, g8, b8)

			index := out.PixOffset(x, y)
			out.Pix[index+0] = r8
			out.Pix[index+1] = g8
			out.Pix[index+2] = b8
			out.Pix[index+3] = alpha
		}
	}

	return out
}

// 計算單一像素在 Chroma Key 後應使用的 alpha 值
//
// 此函式負責：
//  1. 計算像素與 Key Color 的 RGB 色彩距離平方
//  2. 依 Softness 決定使用硬邊或柔邊 keying
//  3. 回傳輸出 PNG 的 alpha 值
//
// 輸入：
//   - params：Key Color、Tolerance、Softness
//   - toleranceSquared：預先計算的 Tolerance²，避免每個像素重複相乘
//   - keyR/keyG/keyB：預先轉為 float64 的背景色
//   - r8/g8/b8：目前來源像素的 8-bit RGB 值
//
// 回傳：
//   - 0：完全透明，背景
//   - 1～254：半透明，柔邊過渡區
//   - 255：完全不透明，前景
func keyAlpha(params KeyParams, toleranceSquared float64, keyR, keyG, keyB float64, r8, g8, b8 uint8) uint8 {

	dr := float64(r8) - keyR
	dg := float64(g8) - keyG
	db := float64(b8) - keyB

	distanceSquared := dr*dr + dg*dg + db*db

	if params.Softness <= 0 {
		return hardKeyAlpha(distanceSquared, toleranceSquared)
	}

	return softKeyAlpha(params, distanceSquared)
}

// 根據色彩距離平方計算硬邊 Chroma Key 的透明度，適合沒有柔化邊緣的硬切去背模式
//
// 參數：
//   - distanceSquared 是目前像素與 Key Color 的 RGB 距離平方：(R - KeyR)² + (G - KeyG)² + (B - KeyB)²
//   - toleranceSquared 是容差的平方：Tolerance²
//
// 回傳值：
//   - 0：目前像素屬於背景，完全透明
//   - 255：目前像素屬於前景，完全不透明
func hardKeyAlpha(distanceSquared float64, toleranceSquared float64) uint8 {

	if distanceSquared <= toleranceSquared {
		return 0
	}

	return 255
}

// 根據色彩距離平方計算柔邊 Chroma Key 的透明度
//
// 它會將透明度分成三段：
//   - distance <= Tolerance => alpha = 0，完全透明
//   - distance >= Tolerance + Softness => alpha = 255，完全不透明
//   - 介於兩者之間 => alpha 由 0 線性漸變到 255
//
// Softness 代表邊緣的過渡寬度；數值越大，主體邊緣越柔和
func softKeyAlpha(params KeyParams, distanceSquared float64) uint8 {

	if params.Softness <= 0 {
		return hardKeyAlpha(distanceSquared, params.Tolerance*params.Tolerance)
	}

	distance := math.Sqrt(distanceSquared)

	start := params.Tolerance
	end := params.Tolerance + params.Softness

	switch {
	case distance <= start:
		return 0
	case distance >= end:
		return 255
	default:
		progress := (distance - start) / (end - start)
		return uint8(progress * 255)
	}
}

// 從輸入串流讀取圖片，執行 Chroma Key 去背後，將含透明度的 PNG 結果寫入輸出串流
//   - 不論輸入是 JPEG 或 PNG，輸出都固定使用 PNG
//   - 因為 PNG 支援 alpha channel，能保存去背後的透明背景
func processReaderWriter(reader io.Reader, writer io.Writer, params KeyParams) error {

	src, _, err := image.Decode(reader)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	out := keyImage(src, params)

	if err := png.Encode(writer, out); err != nil {
		return fmt.Errorf("encode PNG: %w", err)
	}

	return nil
}

// 將角落取樣大小修正為圖片可安全使用的範圍
//   - 回傳值保證符合： 1 <= sampleSize <= min(imageWidth, imageHeight)
func safeSampleSize(bounds image.Rectangle, sampleSize int) (int, error) {

	width := bounds.Dx()
	height := bounds.Dy()

	if width <= 0 || height <= 0 {
		return 0, fmt.Errorf("圖片尺寸不合法")
	}

	if sampleSize <= 0 {
		return 1, nil
	}

	maxSampleSize := min(width, height)

	if sampleSize > maxSampleSize {
		return maxSampleSize, nil
	}

	return sampleSize, nil
}

// 回傳圖片四個角落取樣區塊的左上座標
//
// 回傳順序：左上、右上、左下、右下
func cornersCoordinate(bounds image.Rectangle, sampleSize int) [4][2]int {

	corners := [4][2]int{
		{bounds.Min.X, bounds.Min.Y},
		{bounds.Max.X - sampleSize, bounds.Min.Y},
		{bounds.Min.X, bounds.Max.Y - sampleSize},
		{bounds.Max.X - sampleSize, bounds.Max.Y - sampleSize},
	}

	return corners
}

// 計算圖片多個取樣區塊的平均 RGB 顏色
//   - 此函式會針對 corners 指定的四個角落取樣起點，各自讀取 sampleSize × sampleSize 的像素區域，將所有像素的紅、綠、藍通道值加總後取平均。
func averageColor(img image.Image, sampleSize int, corners [4][2]int) (r, g, b uint8) {

	var sumR, sumG, sumB uint64
	var count int

	bounds := img.Bounds()

	for _, corner := range corners {

		cornerX, cornerY := corner[0], corner[1]

		for dy := range sampleSize {

			for dx := range sampleSize {

				x := cornerX + dx
				y := cornerY + dy

				if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
					continue
				}

				pixel := img.At(x, y)
				r16, g16, b16, _ := pixel.RGBA()

				sumR += uint64(r16 >> 8)
				sumG += uint64(g16 >> 8)
				sumB += uint64(b16 >> 8)

				count++
			}
		}
	}

	if count == 0 {
		return 0, 0, 0
	}

	return uint8(sumR / uint64(count)), uint8(sumG / uint64(count)), uint8(sumB / uint64(count))
}

// parseHexColor 解析 CSS / HTML 十六進位色碼，回傳不透明的 RGBA 顏色
//
// 支援兩種格式：
//   - #RGB：短格式，例如 #0F0，等同於 #00FF00
//   - #RRGGBB：完整格式，例如 #00FF00
//
// 若色碼格式不正確或含有非十六進位字元，回傳 os.ErrInvalid
func parseHexColor(hex string) (color.RGBA, error) {

	rgba, err := parseShortHexColor(hex)
	if err == nil {
		return rgba, nil
	}

	rgba, err = parseFullHexColor(hex)
	if err == nil {
		return rgba, nil
	}

	return color.RGBA{}, os.ErrInvalid
}

// 解析短格式 CSS 色碼：#RGB，將 4-bit 值擴展為 8-bit 值
//
// 例如：
//   - #F00 -> #FF0000
//   - #0F0 -> #00FF00
//   - #00F -> #0000FF
func parseShortHexColor(hex string) (color.RGBA, error) {

	if len(hex) != 4 || hex[0] != '#' {
		return color.RGBA{}, os.ErrInvalid
	}

	red, err := hexToByte(hex[1])
	if err != nil {
		return color.RGBA{}, err
	}

	green, err := hexToByte(hex[2])
	if err != nil {
		return color.RGBA{}, err
	}

	blue, err := hexToByte(hex[3])
	if err != nil {
		return color.RGBA{}, err
	}

	return color.RGBA{R: red<<4 | red, G: green<<4 | green, B: blue<<4 | blue, A: 255}, nil
}

// 解析完整 CSS 色碼：#RRGGBB
//
// 例如：
//   - #FF0000 -> R=255, G=0, B=0
//   - #00FF00 -> R=0, G=255, B=0
//   - #0000FF -> R=0, G=0, B=255
//
// 每一個顏色通道由兩個十六進位字元組成。
// 第一個 nibble 左移 4 bits，第二個 nibble 放在低 4 bits。
func parseFullHexColor(hex string) (color.RGBA, error) {

	if len(hex) != 7 || hex[0] != '#' {
		return color.RGBA{}, os.ErrInvalid
	}

	r1, err := hexToByte(hex[1])
	if err != nil {
		return color.RGBA{}, err
	}

	r2, err := hexToByte(hex[2])
	if err != nil {
		return color.RGBA{}, err
	}

	g1, err := hexToByte(hex[3])
	if err != nil {
		return color.RGBA{}, err
	}

	g2, err := hexToByte(hex[4])
	if err != nil {
		return color.RGBA{}, err
	}

	b1, err := hexToByte(hex[5])
	if err != nil {
		return color.RGBA{}, err
	}

	b2, err := hexToByte(hex[6])
	if err != nil {
		return color.RGBA{}, err
	}

	return color.RGBA{R: r1<<4 | r2, G: g1<<4 | g2, B: b1<<4 | b2, A: 255}, nil
}

// hexToByte 將單一十六進位字元轉換為對應的數值（0～15）
//
// 可接受的字元包括：
//
//	'0' ~ '9'  -> 0 ~ 9
//	'a' ~ 'f'  -> 10 ~ 15
//	'A' ~ 'F'  -> 10 ~ 15
//
// @param char 要轉換的單一 ASCII 十六進位字元
// @return 第一個回傳值為 0～15 的數值；若字元無效則回傳 os.ErrInvalid
func hexToByte(char byte) (byte, error) {

	switch {
	case '0' <= char && char <= '9':
		return char - '0', nil
	case 'a' <= char && char <= 'f':
		return char - 'a' + 10, nil
	case 'A' <= char && char <= 'F':
		return char - 'A' + 10, nil
	}

	return 0, os.ErrInvalid
}
