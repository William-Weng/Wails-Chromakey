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
  let detectedColor: ColorResult = { r: 0, g: 255, b: 0, hex: "#00FF00" };
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
      dialog("info", statusText, outputFileName());
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
      statusText = "去背失敗";
      dialog("error", statusText, error);
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
