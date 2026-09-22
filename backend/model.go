package backend

// 表示從圖片背景取樣後得到的顏色資訊
//
// 例如前端收到的資料會是： {"r":0,"g":255,"b":0,"hex":"#00FF00"}
type ColorResult struct {
	R   uint8  `json:"r"`   // R 是紅色（Red）通道值，範圍為 0 到 255
	G   uint8  `json:"g"`   // G 是綠色（Green）通道值，範圍為 0 到 255
	B   uint8  `json:"b"`   // B 是藍色（Blue）通道值，範圍為 0 到 255
	Hex string `json:"hex"` // Hex 是方便前端顯示與 CSS 使用的十六進位色碼
}

// 定義 Chroma Key 去背演算法的輸入參數
//
// 例如前端傳來的 JSON 可能是： {"keyR":0,"keyG":255,"keyB":0,"tolerance":16,"softness":32}
type KeyParams struct {
	KeyR      uint8   `json:"keyR"`      // KeyR 是要移除的 Key Color 的紅色（Red）通道值
	KeyG      uint8   `json:"keyG"`      // KeyG 是要移除的 Key Color 的綠色（Green）通道值
	KeyB      uint8   `json:"keyB"`      // KeyB 是要移除的 Key Color 的藍色（Blue）通道值
	Tolerance float64 `json:"tolerance"` // Tolerance 是顏色容差, 數值越大，可去除的顏色範圍越廣，但也越可能誤刪前景中相似的顏色
	Softness  float64 `json:"softness"`  // Softness 是邊緣柔化範圍。數值越大，邊緣透明過渡越柔和，但也可能造成主體邊緣看起來較透明
}
