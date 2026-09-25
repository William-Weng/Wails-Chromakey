# [Wails-Chromakey](https://v3.wails.io/features/drag-and-drop/files/)

![Go](https://img.shields.io/badge/Go-1.27.1-00ADD8?logo=go&logoColor=white)
![Node.js](https://img.shields.io/badge/Node.js-20.19.2-339933?logo=nodedotjs&logoColor=white)
![Wails](https://img.shields.io/badge/Wails-v3.0.0--beta.20-DF0000?logo=wails&logoColor=white)
![LICENSE](https://img.shields.io/github/license/William-Weng/Wails-Chromakey?style=flat&label=LICENSE&color=yellow)
![Tag](https://img.shields.io/github/v/tag/William-Weng/Wails-Chromakey?style=flat&label=Tag)
![Stars](https://img.shields.io/github/stars/William-Weng/Wails-Chromakey?style=flat&label=Stars)

一個使用 **Wails 3**、**Go** 與 **Svelte 5** 製作的簡易去背工具。

本專案可以將影片中的指定背景色移除，並輸出為透明背景的 PNG 素材，適合製作綠幕、藍幕或其他純色背景影片素材。

## [操作介面](https://v3.wails.io/zh-tw/reference/cli/)

https://github.com/user-attachments/assets/12e90b7a-07fb-4eb2-ab8f-faddcd435319

## [功能特色](https://v3.wails.io/zh-tw/concepts/build-system/)

- 支援影片去背處理。
- 可調整取樣角落大小。
- 可調整容差，用來控制背景色的辨識範圍。
- 可調整邊緣柔化，降低去背邊緣的鋸齒或色邊。
- 以四角平均色作為 Key Color。
- 輸出 PNG 素材。
- 使用 Wails 3 建立桌面應用程式。

## [技術架構](https://v3.wails.io/zh-tw/tutorials/01-creating-a-service/)

- [Wails 3](https://v3.wails.io/)
- [Go](https://go.dev/)
- [Svelte 5](https://svelte.dev/)
- [Less](https://lesscss.org/)

## 開始使用

### 安裝前端相依套件

進入 `frontend` 目錄，安裝 Less：

```bash
cd frontend
npm install -D less
```

如果專案尚未安裝其他前端相依套件，也可以先執行：

```bash
npm install
```

### 開發模式

請依照 Wails 3 專案設定啟動開發環境：

```bash
wails3 dev
```

開發期間修改 Go、Svelte 或樣式檔案後，Wails 會重新建置或更新應用程式。

## [建置指令整理](https://v3.wails.io/zh-tw/guides/build/building/)

| 目的 | 指令 |
| --- | --- |
| 建立新專案 | `wails3 init -n <專案名稱> -t <前端框架>` |
| 產生 bindings | `wails3 generate bindings` |
| 更新建置資源 | `wails3 task common:update:build-assets` |
| 拉取（下載）用於跨平台交叉編譯的 Docker 映像檔 | `wails3 task setup:docker` |
| 建置 Windows x64 | `wails3 build GOOS=windows GOARCH=amd64` |
| 打包 macOS arm64 | `wails3 package GOOS=darwin GOARCH=arm64` |
| 打包 Linux arm64 | `wails3 build GOOS=linux GOARCH=arm64` |

## 專案結構

```text
.
├── backend/        # golang 後端專案
├── frontend/       # Svelte 5 前端專案
├── app.go          # 應用程式主要邏輯
├── go.mod          # Go 模組設定
└── README.md
```

實際檔案結構可能會依照 Wails 3 專案配置有所不同。

## 除錯

- 在開發模式下，可以使用原系統的瀏覽器工具Debug。

<img width="720" alt="Debug" src="https://github.com/user-attachments/assets/084721e0-ab7b-4c53-ac69-ce0f0dde38c0" />

## 使用流程

1. 開啟應用程式。
2. 輸入來源影片的絕對路徑。
3. 調整取樣角落大小、容差與邊緣柔化參數。
4. 確認顯示的 Key Color 是否正確。
5. 按下「開始去背」。
6. 等待處理完成並取得輸出檔案。

## 注意事項

- 建議使用背景顏色均勻、光線穩定的影片素材。
- 背景若存在陰影、反光或顏色不均，可能需要調整容差與邊緣柔化參數。
- 輸入路徑必須是應用程式執行環境可以存取的檔案路徑。
- 跨平台建置時，請確認目標平台的 Go、Wails 3 與相關建置工具已正確設定。
