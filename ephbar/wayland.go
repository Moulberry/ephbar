package main

import (
	"github.com/rajveermalviya/go-wayland/wayland/client"
	wl_client "github.com/rajveermalviya/go-wayland/wayland/client"
	wlr_layer_shell "github.com/rajveermalviya/go-wayland/wayland/unstable/wlr-layer-shell-v1"
)

type interfaces struct {
	shm        *wl_client.Shm
	compositor *wl_client.Compositor
	seat       *wl_client.Seat
	layerShell *wlr_layer_shell.LayerShell

	registered_count uint8
}

type WaylandState struct {
	display    *wl_client.Display
	registry   *wl_client.Registry
	interfaces *interfaces
}

type output struct {
	wlOutput *wl_client.Output
}

func NewWaylandState() *WaylandState {
	// Connect to wayland server
	display, err := wl_client.Connect("")
	if err != nil {
		fatal("unable to connect to wayland server: %v", err)
	}

	// Get global interfaces registry
	registry, err := display.GetRegistry()
	if err != nil {
		fatal("unable to get global registry object: %v", err)
	}

	return &WaylandState{display: display, registry: registry}
}

func (wayland *WaylandState) context() *client.Context {
	return wayland.display.Context()
}

func (wayland *WaylandState) dispatch() {
	wayland.display.Context().Dispatch()
}

func (wayland *WaylandState) displayRoundtrip() {
	callback, err := wayland.display.Sync()
	if err != nil {
		fatal("unable to get sync callback: %v", err)
	}
	defer callback.Destroy()

	done := false
	callback.AddDoneHandler(func(_ client.CallbackDoneEvent) {
		done = true
	})

	for !done {
		wayland.dispatch()
	}
}

func (wayland *WaylandState) RegisterGlobals() {
	wayland.registry.AddGlobalHandler(wayland.HandleRegistryGlobal)

	// Wait for interfaces
	wayland.displayRoundtrip()

	// Wait for interfaces' events (eg. wl_seat's SeatCapabilitiesEvent)
	wayland.displayRoundtrip()

}

func (wayland *WaylandState) HandleRegistryGlobal(event client.RegistryGlobalEvent) {
	switch event.Interface {
	case "wl_compositor":
		wayland.registerCompositor(event)
	default:
		verbose("discovered unknown interface: %q", event.Interface)
	}
}

func (wayland *WaylandState) registerCompositor(event client.RegistryGlobalEvent) {
	compositor := client.NewCompositor(wayland.context())
	err := wayland.registry.Bind(event.Name, event.Interface, event.Version, compositor)
	if err != nil {
		fatal("unable to bind wl_compositor interface: %v", err)
	}
	if wayland.interfaces.compositor == nil {
		wayland.interfaces.registered_count++
	}
	wayland.interfaces.compositor = compositor
}
