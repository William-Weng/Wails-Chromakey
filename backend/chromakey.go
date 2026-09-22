package backend

import (
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"

	_ "image/jpeg"
)

// SampleCornerColor 從圖片四個角取樣平均，回傳 key color
// sampleSize: 每個角取樣的正方形邊長（像素）
func sampleCornerColor(img image.Image, sampleSize int) (r, g, b uint8) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w <= 0 || h <= 0 {
		return 0, 0, 0
	}

	// 避免 sampleSize 大於圖片尺寸
	if sampleSize > w {
		sampleSize = w
	}
	if sampleSize > h {
		sampleSize = h
	}
	if sampleSize <= 0 {
		sampleSize = 1
	}

	var sumR, sumG, sumB uint64
	var count int

	// 四個角的左上座標
	corners := [][2]int{
		{bounds.Min.X, bounds.Min.Y},                           // 左上
		{bounds.Max.X - sampleSize, bounds.Min.Y},              // 右上
		{bounds.Min.X, bounds.Max.Y - sampleSize},              // 左下
		{bounds.Max.X - sampleSize, bounds.Max.Y - sampleSize}, // 右下
	}

	for _, c := range corners {
		cx, cy := c[0], c[1]
		for dy := 0; dy < sampleSize; dy++ {
			for dx := 0; dx < sampleSize; dx++ {
				x := cx + dx
				y := cy + dy
				if x < bounds.Min.X || x >= bounds.Max.X ||
					y < bounds.Min.Y || y >= bounds.Max.Y {
					continue
				}

				c := img.At(x, y)
				r16, g16, b16, _ := c.RGBA()
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

	return uint8(sumR / uint64(count)),
		uint8(sumG / uint64(count)),
		uint8(sumB / uint64(count))
}

// ProcessFile 讀入圖片，做 Chroma Key，輸出透明 PNG
func processFile(inputPath, outputPath string, p KeyParams) error {
	f, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return err
	}

	out := keyImage(src, p)

	outF, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outF.Close()

	return png.Encode(outF, out)
}

// ProcessReaderWriter 從 io.Reader 讀，寫到 io.Writer
func processReaderWriter(r io.Reader, w io.Writer, p KeyParams) error {
	src, _, err := image.Decode(r)
	if err != nil {
		return err
	}

	out := keyImage(src, p)
	return png.Encode(w, out)
}

// ProcessFileAutoKey 自動從四個角取樣當 key color，然後去背
// sampleSize: 每個角取樣區域的邊長（像素）
func processFileAutoKey(inputPath, outputPath string, tolerance, softness float64, sampleSize int) error {

	f, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return err
	}

	kr, kg, kb := sampleCornerColor(src, sampleSize)

	p := KeyParams{
		KeyR:      kr,
		KeyG:      kg,
		KeyB:      kb,
		Tolerance: tolerance,
		Softness:  softness,
	}

	out := keyImage(src, p)

	outF, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outF.Close()

	return png.Encode(outF, out)
}

func keyImage(src image.Image, p KeyParams) *image.NRGBA {
	bounds := src.Bounds()
	out := image.NewNRGBA(bounds)

	keyRf := float64(p.KeyR)
	keyGf := float64(p.KeyG)
	keyBf := float64(p.KeyB)

	tol2 := p.Tolerance * p.Tolerance

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := src.At(x, y)
			r, g, b, _ := c.RGBA()
			rf := float64(r >> 8)
			gf := float64(g >> 8)
			bf := float64(b >> 8)

			dr := rf - keyRf
			dg := gf - keyGf
			db := bf - keyBf
			d2 := dr*dr + dg*dg + db*db

			var alpha uint8
			if p.Softness <= 0 {
				if d2 <= tol2 {
					alpha = 0
				} else {
					alpha = 255
				}
			} else {
				d := math.Sqrt(d2)
				t0 := p.Tolerance
				t1 := p.Tolerance + p.Softness
				if d <= t0 {
					alpha = 0
				} else if d >= t1 {
					alpha = 255
				} else {
					a := (d - t0) / (t1 - t0)
					alpha = uint8(a * 255)
				}
			}

			i := out.PixOffset(x, y)
			out.Pix[i+0] = uint8(rf + 0.5)
			out.Pix[i+1] = uint8(gf + 0.5)
			out.Pix[i+2] = uint8(bf + 0.5)
			out.Pix[i+3] = alpha
		}
	}

	return out
}

// 小工具：從 hex 字串 (#00FF00) 轉 KeyParams
func FromHex(hexColor string, tolerance, softness float64) (KeyParams, error) {
	c, err := parseHexColor(hexColor)
	if err != nil {
		return KeyParams{}, err
	}
	return KeyParams{
		KeyR:      c.R,
		KeyG:      c.G,
		KeyB:      c.B,
		Tolerance: tolerance,
		Softness:  softness,
	}, nil
}

func parseHexColor(s string) (color.RGBA, error) {
	if len(s) == 4 && s[0] == '#' {
		r, err := hexToByte(s[1])
		if err != nil {
			return color.RGBA{}, err
		}
		g, err := hexToByte(s[2])
		if err != nil {
			return color.RGBA{}, err
		}
		b, err := hexToByte(s[3])
		if err != nil {
			return color.RGBA{}, err
		}
		return color.RGBA{R: r, G: g, B: b, A: 255}, nil
	}
	if len(s) == 7 && s[0] == '#' {
		r1, _ := hexToByte(s[1])
		r2, _ := hexToByte(s[2])
		g1, _ := hexToByte(s[3])
		g2, _ := hexToByte(s[4])
		b1, _ := hexToByte(s[5])
		b2, _ := hexToByte(s[6])
		return color.RGBA{
			R: r1<<4 | r2,
			G: g1<<4 | g2,
			B: b1<<4 | b2,
			A: 255,
		}, nil
	}
	return color.RGBA{}, os.ErrInvalid
}

func hexToByte(c byte) (byte, error) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', nil
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, nil
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, nil
	}
	return 0, os.ErrInvalid
}
