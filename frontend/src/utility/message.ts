  /**
   * 將 unknown error 轉成可顯示的文字
   *
   * @param error - try/catch 捕捉到的任意錯誤值
   * @returns 可安全顯示在 UI 或 dialog 的錯誤訊息
   */
  export function errorMessage(error: unknown): string {
    return error instanceof Error ? error.message : String(error);
  }