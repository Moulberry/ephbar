package main

import (
	"github.com/rajveermalviya/go-wayland/wayland/client"
)

func (seat *Seat) attachKeyboard() {
	keyboard, err := seat.wlSeat.GetKeyboard()
	if err != nil {
		fatal("unable to register keyboard interface")
	}
	seat.keyboard = keyboard

	app := seat.wayland.appState
	keyboard.AddKeyHandler(app.HandleKeyboardKey)

	info("keyboard interface registered")
}

func (seat *Seat) releaseKeyboard() {
	if err := seat.keyboard.Release(); err != nil {
		fatal("unable to release keyboard interface")
	}
	seat.keyboard = nil

	info("keyboard interface released")
}

func (app *AppState) HandleKeyboardKey(e client.KeyboardKeyEvent) {
	// close on "esc"
	if e.Key == 1 {
		app.exit = true
	}
}
