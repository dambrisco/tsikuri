# Development Guide

## Prerequisites

- Go 1.21+
- Linux: `xdotool`, `xclip`, ImageMagick (`import` command), X11 display

## Verify changes before pushing

1. **Format**: `gofmt -l .` — should produce no output
2. **Vet**: `go vet ./...`
3. **Test**: `go test -race ./...`
4. **Build (Linux)**: `GOOS=linux GOARCH=amd64 go build ./...`
5. **Build (Windows)**: `GOOS=windows GOARCH=amd64 go build ./...`

## Project structure

| Package | Description |
|---|---|
| `tsikuri` | Core API — Screen, Region, Pattern, Match, Observer, interfaces |
| `match` | Pure-Go NCC template matching (implements `tsikuri.Matcher`) |
| `platform` | OS-native screen capture, input simulation, OCR |
| `tsikuritest` | Mock backends for unit testing |
