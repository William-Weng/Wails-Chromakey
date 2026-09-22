/**
 * 取得檔案路徑中的檔名部分
 *
 * 同時支援：
 * - macOS / Linux 路徑：`/Users/ios/Desktop/images.jpg`
 * - Windows 路徑：`C:\Users\ios\Desktop\images.jpg`
 *
 * @param path - 完整檔案路徑或單純檔名。
 * @returns 檔名，例如 `images.jpg`。
 */
export function fileNameFromPath(path: string): string {
    return path.split(/[\\/]/).pop() ?? path;
}

/**
 * 建立目前時間的 Unix timestamp 字串
 *
 * Date.now() 回傳從 Unix Epoch 開始計算的毫秒數，適合作為輸出檔名的一部分，降低重複檔名機率
 *
 * @returns 毫秒級 Unix timestamp 字串。
 */
export function timestamp(): string {
    return String(Date.now());
}

/**
 * 根據來源圖片路徑產生去背後的 PNG 輸出路徑
 *
 * 規則：
 *
 * ```text
 * 原始路徑：/Users/ios/Desktop/images.jpg
 * 輸出路徑：/Users/ios/Desktop/images-1790047081234.png
 * ```
 *
 * @param inputPath - 原始圖片檔案的完整路徑。
 * @returns 自動產生且固定為 `.png` 的輸出檔案路徑。
 */
export function chromaKeyOutputPath(inputPath: string): string {

    const lastSlashIndex = Math.max(inputPath.lastIndexOf("/"), inputPath.lastIndexOf("\\"));
    const lastDotIndex = inputPath.lastIndexOf(".");
    const hasExtension = lastDotIndex > lastSlashIndex;
    const basePath = hasExtension ? inputPath.slice(0, lastDotIndex) : inputPath;

    return `${basePath}-${timestamp()}.png`;
}