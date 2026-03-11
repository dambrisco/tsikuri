// Package tsikuri provides image-based GUI automation for Go.
//
// Tsikuri allows you to find images on screen and interact with them
// programmatically — clicking buttons, typing text, waiting for UI
// elements to appear or disappear, and more. It is inspired by
// SikuliX but designed as an idiomatic Go library.
//
// The library uses pluggable backends for screen capture, template
// matching, input simulation, and OCR, with platform-native default
// implementations that produce self-contained binaries with no
// external shared library dependencies.
//
// Basic usage:
//
//	screen, err := tsikuri.NewScreen(
//	    tsikuri.WithBasePath("./images"),
//	    tsikuri.WithTimeout(5*time.Second),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	ctx := context.Background()
//	err = screen.Click(ctx, tsikuri.Image("submit.png"))
package tsikuri
