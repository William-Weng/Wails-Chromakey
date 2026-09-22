package backend

type ColorResult struct {
	R   uint8  `json:"r"`
	G   uint8  `json:"g"`
	B   uint8  `json:"b"`
	Hex string `json:"hex"`
}

type KeyParams struct {
	KeyR      uint8   `json:"keyR"`
	KeyG      uint8   `json:"keyG"`
	KeyB      uint8   `json:"keyB"`
	Tolerance float64 `json:"tolerance"`
	Softness  float64 `json:"softness"`
}
