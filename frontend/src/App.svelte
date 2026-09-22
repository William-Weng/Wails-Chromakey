<script lang="ts">
  import { onMount } from "svelte";
  import { Events } from "@wailsio/runtime";
  import { dialog } from "./utility/dialog";
  import { eventData } from "./utility/event";

  import {
    CornerColorFromFile,
    ProcessFile,
  } from "../bindings/wails-chromakey/backend/ChromakeyService";

  let inputPath = "";
  let outputPath = "";
  let sampleSize = 16;
  let tolerance = 16;
  let softness = 16;
  let detectedColor: ColorResult = {r:0, g:255, b:0, hex: "#00FF00"};
  let processing = false;
  let error = "";
  let statusText = "";

  function makeOutputPath(path: string): string {
    const slash = Math.max(path.lastIndexOf("/"), path.lastIndexOf("\\"));
    const dot = path.lastIndexOf(".");
    const base = dot > slash ? path.slice(0, dot) : path;
    const timestamp = makeTimestamp();

    return `${base}-${timestamp}.png`;
  }

  function makeTimestamp(): string {
    return Date.now().toString();
  }

  function outputFileName(): string {
    return outputPath.split(/[\\/]/).pop() ?? outputPath;
  }

  async function detectCornerColor(): Promise<void> {
    const path = inputPath.trim();

    if (!path) {
      error = "請先輸入圖片的完整絕對路徑。";
      return;
    }

    processing = true;
    error = "";
    detectedColor = null;
    statusText = "正在偵測背景色…";

    try {
      detectedColor = await CornerColorFromFile(path, sampleSize);
      statusText = "背景色偵測完成";
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
      statusText = "背景色偵測失敗";
    } finally {
      processing = false;
    }
  }

  async function processImage(): Promise<void> {
    
    const source = inputPath.trim();
    outputPath = makeOutputPath(source);
    const destination = outputPath.trim();

    if (!source) {
      error = "請先輸入來源圖片路徑。";
      return;
    }

    if (!destination) {
      error = "輸出路徑不可為空。";
      return;
    }

    if (!detectedColor) {
      error = "請先偵測背景色。";
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
      dialog("info", statusText, outputFileName())
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
      statusText = "去背失敗";
      dialog("error", statusText, error)
    } finally {
      processing = false;
    }
  }

  onMount(() => {
    const unsubscribeDrop = Events.On("image-file-dropped", (event) => {
      void imageFileDroppedAction(event);
    });

    return () => {
      unsubscribeDrop();
    };
  });

  async function imageFileDroppedAction(event: unknown): Promise<void> {
    const files = eventData<string[]>(event);

    if (!Array.isArray(files) || files.length === 0) {
      error = "沒有收到有效的圖片檔案。";
      return;
    }

    inputPath = files[0];
    outputPath = "";
    error = "";

    await detectCornerColor();
  }

</script>

<svelte:head>
  <title>綠幕去背小工具</title>
</svelte:head>

<main class="page">
  <section class="card" data-file-drop-target>
    <label class="field">
      <span>來源圖片絕對路徑</span>
      <input
        bind:value={inputPath}
        placeholder="/Users/ios/Desktop/images.jpg"
      />
    </label>

      <section class="color-result">
        <div
          class="swatch"
          style={`background-color: ${detectedColor.hex}`}
        ></div>

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

    <button
      class="primary-button"
      type="button"
      disabled={processing || !detectedColor}
      on:click={processImage}
    >
      {processing ? "處理中…" : "開始去背並輸出 PNG"}
    </button>

  </section>
</main>

<style>
  /*
 * Chroma Key layout reset
 *
 * 目標：
 * 1. 頁面四邊貼齊視窗，不保留瀏覽器預設空白。
 * 2. 主卡片使用滿版寬度。
 * 3. 所有 range 共用相同寬度。
 * 4. 內容仍保留合理的內距，不讓文字貼住視窗邊緣。
 */

:global(*) {
  box-sizing: border-box;
}

:global(html),
:global(body),
:global(#app) {
  width: 100%;
  min-width: 0;
  min-height: 100%;
  margin: 0;
  padding: 0;
}

:global(body) {
  overflow-x: hidden;
  background: #f5f6fb;
  color: #536174;
  font-family: Inter, ui-sans-serif, system-ui, -apple-system,
    BlinkMacSystemFont, "Segoe UI", sans-serif;
}

/* 頁面四邊不留白；padding 只作為內容的內縮距離 */
.page {
  width: 100%;
  min-height: 100vh;
  padding: 0;
}

/* 主卡片滿版，不設定 max-width */
.card {
  width: 100%;
  min-height: 100vh;
  margin: 0;
  padding: 30px 48px 36px;
  border: 0;
  border-radius: 0;
  background: #fff;
  box-shadow: none;
}

.field,
.slider-field {
  width: 100%;
  display: grid;
  gap: 10px;
}

.field > span,
.slider-field > span {
  color: #5c6b7f;
  font-size: 16px;
  font-weight: 800;
}

.field input {
  width: 100%;
  height: 52px;
  padding: 0 17px;
  border: 1px solid #bdc7d5;
  border-radius: 11px;
  outline: none;
  background: #202020;
  color: #f3f4f6;
  font-size: 15px;
}

.field input:focus {
  border-color: #3478f6;
  box-shadow: 0 0 0 3px rgba(52, 120, 246, 0.14);
}

.color-result {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 20px;
  margin-top: 22px;
  padding: 22px 20px;
  border: 1px solid #dce4ef;
  border-radius: 14px;
  background: #f9fbfe;
}

.swatch {
  width: 74px;
  height: 74px;
  flex: 0 0 auto;
  border: 1px solid #94b2ce;
  border-radius: 13px;
}

.color-info {
  display: grid;
  gap: 4px;
}

.color-info small {
  color: #8a99ad;
  font-size: 16px;
}

.color-info strong {
  color: #263548;
  font-size: 27px;
  letter-spacing: 0.4px;
}

.color-info span {
  color: #8a99ad;
  font-size: 16px;
}

/* 三個滑桿統一滿寬 */
.slider-field {
  margin-top: 24px;
}

input[type="range"] {
  display: block;
  width: 100%;
  min-width: 0;
  height: 22px;
  margin: 0;
  padding: 0;
  accent-color: #8e8997;
  cursor: pointer;
}

.primary-button {
  display: block;
  width: 100%;
  height: 53px;
  margin-top: 25px;
  border: 0;
  border-radius: 11px;
  background: #3478f6;
  color: #fff;
  font-size: 16px;
  font-weight: 800;
  cursor: pointer;
}

.primary-button:hover:not(:disabled) {
  background: #2469e8;
}

.primary-button:disabled {
  cursor: wait;
  opacity: 0.55;
}
</style>
