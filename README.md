# tsikuri

Image-based GUI automation for Go, inspired by [SikuliX](http://sikulix.com/).

Find images on screen and interact with them — click buttons, type text, wait
for UI elements to appear or disappear — all from compiled Go binaries with no
external shared library dependencies.

## Features

- **Find & click** — locate a UI element by its screenshot and click it
- **Wait & react** — wait for images to appear/disappear with context-based timeouts
- **Observer** — monitor a region for visual changes in the background
- **Sub-regions** — restrict operations to a portion of the screen
- **OCR** — extract text from screen regions (Windows, via WinRT)
- **Cross-platform** — Windows (primary) and Linux (X11) with pluggable backends
- **No external dependencies** — builds to a single static binary; uses OS-native
  APIs (Win32 GDI/SendInput, xdotool) rather than OpenCV or Tesseract

## Install

```
go get github.com/dambrisco/tsikuri
```

### System requirements

**Windows:** None beyond a standard Go toolchain.

**Linux:** `xdotool`, `xclip`, and ImageMagick (`import` command) must be
installed and an X11 display must be available.

## Quick start

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/dambrisco/tsikuri"
)

func main() {
    screen, err := tsikuri.NewScreen(
        tsikuri.WithBasePath("./images"),
        tsikuri.WithTimeout(5*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Find a button on screen and click it
    err = screen.Click(ctx, tsikuri.Image("submit.png"))
    if err != nil {
        log.Fatal(err)
    }
}
```

## Examples

### Wait for an element, then interact

```go
// Wait for a dialog to appear (respects context deadline or AutoWaitTimeout)
match, err := screen.Wait(ctx, tsikuri.Image("login_dialog.png"))
if err != nil {
    log.Fatal("dialog never appeared:", err)
}

// Click the match, then type into it
match.Click(ctx)
match.Type("username")
match.Type("\t") // tab to next field
match.Paste("hunter2")

// Submit and wait for the dialog to go away
screen.Click(ctx, tsikuri.Image("login_button.png"))
screen.WaitVanish(ctx, tsikuri.Image("login_dialog.png"))
```

### Work within a sub-region

```go
// Only search the left 400 pixels of the screen
sidebar := screen.Left(400)
sidebar.Click(ctx, tsikuri.Image("menu_item.png"))

// Search below a matched element
header, _ := screen.Find(ctx, tsikuri.Image("section_header.png"))
content := header.Below(300)
items, _ := content.FindAll(ctx, tsikuri.Image("list_item.png"))
log.Printf("found %d items", len(items))
```

### Adjust match similarity

```go
// Default similarity is 0.7; raise it for pixel-exact matching
exact := tsikuri.Image("icon.png").WithSimilarity(0.95)

// Or lower it to tolerate anti-aliasing / theme differences
fuzzy := tsikuri.Image("button.png").WithSimilarity(0.5)

screen.Click(ctx, exact)
screen.Click(ctx, fuzzy)
```

### Observer — react to screen changes

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

obs := screen.OnAppear(tsikuri.Image("error_popup.png"), func(e tsikuri.ObserveEvent) {
    log.Println("error popup detected, dismissing")
    e.Match.Click(context.Background(), tsikuri.Image("ok_button.png"))
})

go obs.Run(ctx) // runs until ctx is cancelled
```

### Pick the best match from several patterns

```go
match, index, err := screen.FindBest(ctx,
    tsikuri.Image("save_button_v1.png"),
    tsikuri.Image("save_button_v2.png"),
    tsikuri.Image("save_button_dark.png"),
)
if err == nil {
    log.Printf("matched pattern %d with score %.2f", index, match.Score)
    match.Click(ctx)
}
```

## Custom backends

All platform-specific operations go through interfaces defined in the `tsikuri`
package. You can substitute any of them:

```go
screen, _ := tsikuri.NewScreen(
    tsikuri.WithBackends(&tsikuri.Backends{
        Capture: myCustomCapturer,   // implements tsikuri.Capturer
        Match:   match.New(),        // default pure-Go NCC matcher
        Input:   myCustomInput,      // implements tsikuri.Inputter
        OCR:     myCustomOCR,        // implements tsikuri.OCREngine
    }),
)
```

## Testing

The `tsikuritest` package provides mock backends so you can unit-test automation
scripts without a display:

```go
func TestMyWorkflow(t *testing.T) {
    screenImg := loadTestScreenshot("testdata/desktop.png")
    mockCapture := tsikuritest.NewMockCapture(screenImg)
    mockInput := tsikuritest.NewMockInput()

    screen, _ := tsikuri.NewScreen(
        tsikuri.WithBackends(&tsikuri.Backends{
            Capture: mockCapture,
            Match:   match.New(),
            Input:   mockInput,
        }),
        tsikuri.WithBasePath("testdata/"),
    )

    ctx := context.Background()
    screen.Click(ctx, tsikuri.Image("button.png"))

    if !mockInput.AssertClicked(120, 45) {
        t.Error("expected click at button location")
    }
}
```

## Package structure

| Package | Description |
|---|---|
| `tsikuri` | Core API — Screen, Region, Pattern, Match, Observer, interfaces |
| `match` | Pure-Go NCC template matching (implements `tsikuri.Matcher`) |
| `platform` | OS-native screen capture, input simulation, OCR |
| `tsikuritest` | Mock backends for unit testing |
