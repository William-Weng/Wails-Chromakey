
# Wails-Chromakey

一個使用 Wails 3、Go、Svelte 5 製作的去背工具。

- 影片
https://github.com/user-attachments/assets/b7f9e4ae-dab8-4d0a-9313-8fe6e6624e44

- 使用less
```
cd frontend
npm install -D less
```

- 常用Wails指令

```bash
wails3 generate bindings
wails3 task common:update:build-assets
wails3 build GOOS=windows GOARCH=amd64
wails3 package GOOS=darwin GOARCH=arm64
```