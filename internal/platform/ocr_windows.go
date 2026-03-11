//go:build windows

package platform

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// WindowsOCR implements OCRBackend using Windows built-in OCR.
// It shells out to a PowerShell script that uses the Windows.Media.Ocr WinRT API.
type WindowsOCR struct{}

// NewOCR creates a new WindowsOCR backend.
func NewOCR() *WindowsOCR {
	return &WindowsOCR{}
}

func (o *WindowsOCR) ReadText(img image.Image) (string, error) {
	// Write image to a temp file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "tsikuri_ocr.png")

	f, err := os.Create(tmpFile)
	if err != nil {
		return "", fmt.Errorf("tsikuri: failed to create temp file for OCR: %w", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return "", fmt.Errorf("tsikuri: failed to encode image for OCR: %w", err)
	}
	f.Close()
	defer os.Remove(tmpFile)

	// Use PowerShell to invoke Windows OCR
	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Runtime.WindowsRuntime
$null = [Windows.Media.Ocr.OcrEngine,Windows.Foundation,ContentType=WindowsRuntime]
$null = [Windows.Graphics.Imaging.BitmapDecoder,Windows.Foundation,ContentType=WindowsRuntime]

$path = '%s'
$stream = [System.IO.File]::OpenRead($path)
$randomAccessStream = [System.IO.WindowsRuntimeStreamExtensions]::AsRandomAccessStream($stream)

$decoder = [Windows.Graphics.Imaging.BitmapDecoder]::CreateAsync($randomAccessStream)
$decoder = $decoder.GetAwaiter().GetResult()

$bitmap = $decoder.GetSoftwareBitmapAsync()
$bitmap = $bitmap.GetAwaiter().GetResult()

$engine = [Windows.Media.Ocr.OcrEngine]::TryCreateFromUserProfileLanguages()
$result = $engine.RecognizeAsync($bitmap)
$result = $result.GetAwaiter().GetResult()

Write-Output $result.Text

$stream.Close()
`, strings.ReplaceAll(tmpFile, `\`, `\\`))

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tsikuri: Windows OCR failed: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}

func (o *WindowsOCR) Close() error {
	return nil
}
