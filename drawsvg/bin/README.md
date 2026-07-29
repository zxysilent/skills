# drawcli binaries

The bundled binaries are built from `drawsvg/drawcli` with:

```text
go build -trimpath -buildvcs=false -ldflags="-s -w"
```

- `drawcli-linux-amd64` for x86-64 Linux
- `drawcli-linux-arm64` for ARM64 Linux
- `drawcli-windows-amd64.exe` for x86-64 Windows
- `drawcli-darwin-amd64` for Intel macOS
- `drawcli-darwin-arm64` for Apple Silicon macOS

Rebuild them with:

```bash
drawsvg/scripts/build-drawcli.sh
```
