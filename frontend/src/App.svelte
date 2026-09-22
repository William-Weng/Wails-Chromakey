<script lang="ts">
  import { onMount } from "svelte";
  import { Events } from "@wailsio/runtime";

  import { dialog } from "./utility/dialog";
  import { eventData } from "./utility/event";
  import { chromaKeyOutputPath, fileNameFromPath } from "./utility/path";
  
  import { CornerColorFromFile, ProcessFile } from "../bindings/wails-chromakey/backend/ChromakeyService";

  let inputPath = "";
  let outputPath = "";
  let sampleSize = 16;
  let tolerance = 16;
  let softness = 16;
  let detectedColor: ColorResult = { r: 0, g: 255, b: 0, hex: "#00FF00" };
  let processing = false;
  let error = "";
  let statusText = "";

  onMount(() => {
    const unsubscribeDrop = Events.On("image-file-dropped", (event) => {
      void imageFileDroppedAction(event);
    });

    return () => {
      unsubscribeDrop();
    };
  });

  /**
   * 處理 Wails 的 image-file-dropped 事件
   *
   * Go 後端會在使用者將圖片拖入視窗時發送事件：
   *
   * ```go
   * application.Get().Event.Emit("image-file-dropped", files)
   * ```
   *
   * 前端收到的 payload 預期為 string[]，內容是被拖入圖片的本機絕對路徑
   *
   * 處理流程：
   * 1. 從 Wails event 取出檔案路徑陣列
   * 2. 驗證是否至少收到一個檔案
   * 3. 使用第一個檔案作為目前來源圖片
   * 4. 清除上一次圖片的輸出路徑、偵測結果、錯誤與狀態
   * 5. 自動呼叫 Go 後端偵測圖片四角平均背景色
   *
   * @param event - Wails Events.On() callback 收到的事件物件或 payload
   */
  async function imageFileDroppedAction(event: unknown): Promise<void> {
    
    const files = eventData<string[]>(event);

    if (!Array.isArray(files) || files.length === 0) {
      error = "沒有收到有效的圖片檔案。";
      statusText = "拖放圖片失敗";
      await dialog("warning", statusText, error);
      return;
    }

    inputPath = files[0];
    outputPath = "";
    error = "";
    statusText = "已載入圖片";

    await detectCornerColor();
  }

  /**
   * 執行 Chroma Key 去背並輸出透明 PNG
   *
   * 流程：
   * 1. 驗證來源圖片路徑與已偵測的 Key Color
   * 2. 依來源檔案自動建立帶 timestamp 的 PNG 輸出路徑
   * 3. 建立 KeyParams
   * 4. 呼叫 Go / Wails 的 ProcessFile()
   * 5. 以原生系統 dialog 顯示成功或失敗結果
   */
  async function processImage(): Promise<void> {
    const source = inputPath.trim();

    if (!source) {
      error = "請先輸入來源圖片路徑。";
      statusText = "尚未選擇圖片";
      await dialog("warning", statusText, error);
      return;
    }

    if (!detectedColor) {
      error = "請先偵測背景色。";
      statusText = "尚未偵測背景色";
      await dialog("warning", statusText, error);
      return;
    }

    outputPath = chromaKeyOutputPath(source);

    const destination = outputPath.trim();

    if (!destination) {
      error = "輸出路徑不可為空。";
      statusText = "無法建立輸出路徑";
      await dialog("error", statusText, error);
      return;
    }

    processing = true;
    error = "";
    statusText = "正在去背…";

    const params: KeyParams = {
      keyR: detectedColor.r,
      keyG: detectedColor.g,
      keyB: detectedColor.b,
      tolerance,
      softness,
    };

    try {
      await ProcessFile(source, destination, params);
      statusText = "去背完成";
      await dialog("info", statusText, `已輸出 PNG：\n${fileNameFromPath(destination)}`);
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
      statusText = "去背失敗";
      await dialog("error", statusText, error);
    } finally {
      processing = false;
    }
  }

  /**
   * 呼叫 Go 後端，從來源圖片四個角落取樣並取得平均背景色
   *
   * 成功時更新：
   * - detectedColor：四角平均 RGB / Hex 色彩
   * - statusText：背景色偵測完成。
   *
   * 失敗時更新：
   * - error：可顯示的錯誤訊息
   * - statusText：背景色偵測失敗
   * - 原生錯誤 dialog
   */
  async function detectCornerColor(): Promise<void> {
    const path = inputPath.trim();

    if (!path) {
      error = "請先輸入圖片的完整絕對路徑。";
      statusText = "尚未選擇圖片";
      await dialog("warning", statusText, error);
      return;
    }

    processing = true;
    error = "";
    statusText = "正在偵測背景色…";

    try {
      detectedColor = await CornerColorFromFile(path, sampleSize);
      statusText = "背景色偵測完成";
    } catch (err) {
      error = errorMessage(err);
      statusText = "背景色偵測失敗";
      await dialog("error", statusText, error);
    } finally {
      processing = false;
    }
  }

  /**
   * 處理 HTML color input 的顏色變更事件
   *
   * 使用者從調色盤選擇顏色後，會執行這個函式，並同步更新 detectedColor 的 HEX 與 RGB 值
   *
   * @param event color input 的 change 事件
   */ 
  function handleColorChange(event: Event) {

    const input = event.currentTarget as HTMLInputElement;
    const hex = input.value.toUpperCase();
    const rgb = hexToRgb(hex);

    detectedColor = { hex,...rgb };
  }

  /**
   * 將 HEX 色碼轉換成 RGB 色彩值
   *
   * 例如：
   * "#00FF00" 會轉換成： {"r": 0, "g": 255, "b": 0}
   *
   * @param hex HEX 顏色字串，例如 "#00FF00"
   * @returns 包含紅、綠、藍三個色彩通道的物件
   */
  function hexToRgb(hex: string) {

      const value = hex.replace('#', '');

      return {
          r: parseInt(value.slice(0, 2), 16),
          g: parseInt(value.slice(2, 4), 16),
          b: parseInt(value.slice(4, 6), 16)
      };
  }

  /**
   * 將 unknown error 轉成可顯示的文字
   *
   * @param error - try/catch 捕捉到的任意錯誤值
   * @returns 可安全顯示在 UI 或 dialog 的錯誤訊息
   */
  function errorMessage(error: unknown): string {
    return error instanceof Error ? error.message : String(error);
  }
</script>

<svelte:head>
  <title>綠幕去背小工具</title>
</svelte:head>

<main class="page">
  <section class="card" data-file-drop-target>
    <label class="field">
      <span>來源圖片絕對路徑</span>
      <input bind:value={inputPath} placeholder="/Users/your_name/Desktop/images.jpg"/>
    </label>

    <section class="color-result">

    <label
        class="swatch"
        style={`background-color: ${detectedColor.hex}`}
        aria-label="選擇 Key Color"
    >
        <input
            type="color"
            value={detectedColor.hex}
            on:change={handleColorChange}
        />
    </label>

      <div class="color-info">
        <small>四角平均 Key Color</small>
        <strong>{detectedColor.hex}</strong>
        <span>
          RGB({detectedColor.r}, {detectedColor.g}, {detectedColor.b})
        </span>
      </div>
    </section>

    <label class="slider-field">
      <span>角落取樣大小：{sampleSize}px</span>
      <input type="range" min="1" max="100" bind:value={sampleSize} />
    </label>

    <label class="slider-field">
      <span>容差：{tolerance}</span>
      <input type="range" min="0" max="220" bind:value={tolerance} />
    </label>

    <label class="slider-field">
      <span>邊緣柔化：{softness}</span>
      <input type="range" min="0" max="100" bind:value={softness} />
    </label>

    <button class="primary-button" type="button" disabled={processing || !detectedColor} on:click={processImage}>
      {processing ? "處理中…" : "開始去背並輸出 PNG"}
    </button>
  </section>
</main>
