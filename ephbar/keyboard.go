package main

import (
	"github.com/rajveermalviya/go-wayland/wayland/client"
)

func (app *appState) attachKeyboard() {
	keyboard, err := app.seat.GetKeyboard()
	if err != nil {
		fatal("unable to register keyboard interface")
	}
	app.keyboard = keyboard

	keyboard.AddKeyHandler(app.HandleKeyboardKey)

	info("keyboard interface registered")
}

func (app *appState) releaseKeyboard() {
	if err := app.keyboard.Release(); err != nil {
		fatal("unable to release keyboard interface")
	}
	app.keyboard = nil

	info("keyboard interface released")
}

func (app *appState) HandleKeyboardKey(e client.KeyboardKeyEvent) {
	// close on "esc"
	if e.Key == 1 {
		app.exit = true
	}
}
