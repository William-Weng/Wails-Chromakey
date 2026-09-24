
  /**
   * 將 HEX 色碼轉換成 RGB 色彩值
   *
   * 例如：
   * "#00FF00" 會轉換成： {"r": 0, "g": 255, "b": 0}
   *
   * @param hex HEX 顏色字串，例如 "#00FF00"
   * @returns 包含紅、綠、藍三個色彩通道的物件
   */
  export function hexToRgb(hex: string) {

      const value = hex.replace('#', '');

      return {
          r: parseInt(value.slice(0, 2), 16),
          g: parseInt(value.slice(2, 4), 16),
          b: parseInt(value.slice(4, 6), 16)
      };
  }