package main

import (
	"github.com/rajveermalviya/go-wayland/wayland/client"
	wl_client "github.com/rajveermalviya/go-wayland/wayland/client"
	wlr_layer_shell "github.com/rajveermalviya/go-wayland/wayland/unstable/wlr-layer-shell-v1"
)

type interfaces struct {
	shm        *wl_client.Shm
	compositor *wl_client.Compositor
	layerShell *wlr_layer_shell.LayerShell

	registeredCount uint8
}

type WaylandState struct {
	appState *AppState

	display    *wl_client.Display
	registry   *wl_client.Registry
	interfaces interfaces

	outputs []*Output
	seats   []*Seat
}

type Output struct {
	wayland  *WaylandState
	wlOutput *wl_client.Output

	surface *Surface
}

type Surface struct {
	output *Output

	width, height uint32

	wlSurface    *client.Surface
	layerSurface *wlr_layer_shell.LayerSurface
}

type Seat struct {
	wayland *WaylandState
	wlSeat  *wl_client.Seat

	keyboard *client.Keyboard
	pointer  *client.Pointer
}

func NewWaylandState(appState *AppState) *WaylandState {
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
	return &WaylandState{appState: appState, display: display, registry: registry}
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

	if wayland.interfaces.registeredCount != 3 {
		fatal("missing wayland interfaces. is riverwm running?")
	} else {
		verbose("all interfaces registered")
	}
}

// Interface registration

func (wayland *WaylandState) HandleRegistryGlobal(event client.RegistryGlobalEvent) {
	switch event.Interface {
	case "wl_compositor":
		wayland.registerCompositor(event)
	case "wl_shm":
		wayland.registerShm(event)
	case "wl_seat":
		wayland.registerSeat(event)
	case "wl_output":
		wayland.registerOutput(event)
	case "zwlr_layer_shell_v1":
		wayland.registerLayerShell(event)
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
		wayland.interfaces.registeredCount++
	}
	wayland.interfaces.compositor = compositor
}

func (wayland *WaylandState) registerShm(event client.RegistryGlobalEvent) {
	shm := client.NewShm(wayland.context())
	err := wayland.registry.Bind(event.Name, event.Interface, event.Version, shm)
	if err != nil {
		fatal("unable to bind wl_shm interface: %v", err)
	}
	if wayland.interfaces.shm == nil {
		wayland.interfaces.registeredCount++
	}
	wayland.interfaces.shm = shm
}

func (wayland *WaylandState) registerSeat(event client.RegistryGlobalEvent) {
	wlSeat := client.NewSeat(wayland.context())
	err := wayland.registry.Bind(event.Name, event.Interface, event.Version, wlSeat)
	if err != nil {
		fatal("unable to bind wl_seat interface: %v", err)
	}

	seat := &Seat{wayland: wayland, wlSeat: wlSeat}
	wlSeat.AddCapabilitiesHandler(seat.HandleSeatCapabilities)

	wayland.seats = append(wayland.seats, seat)
}

func (wayland *WaylandState) registerOutput(event client.RegistryGlobalEvent) {
	wlOutput := client.NewOutput(wayland.context())
	err := wayland.registry.Bind(event.Name, event.Interface, event.Version, wlOutput)
	if err != nil {
		fatal("unable to bind wl_output interface: %v", err)
	}

	output := &Output{wayland: wayland, wlOutput: wlOutput}
	wlOutput.AddDoneHandler(output.HandleOutputDone)

	wayland.outputs = append(wayland.outputs, output)
}

func (wayland *WaylandState) registerLayerShell(event client.RegistryGlobalEvent) {
	layerShell := wlr_layer_shell.NewLayerShell(wayland.context())
	err := wayland.registry.Bind(event.Name, event.Interface, event.Version, layerShell)
	if err != nil {
		fatal("unable to bind wl_layer_shell interface: %v", err)
	}
	if wayland.interfaces.layerShell == nil {
		wayland.interfaces.registeredCount++
	}
	wayland.interfaces.layerShell = layerShell
}

// Handlers

func (output *Output) HandleOutputDone(e client.OutputDoneEvent) {
	if output.surface == nil {
		output.createSurface(50)
	}
}

func (output *Output) createSurface(height uint32) {
	wayland := output.wayland
	interfaces := wayland.interfaces

	// Create a wlSurface for toplevel window
	wlSurface, err := interfaces.compositor.CreateSurface()
	if err != nil {
		fatal("unable to create compositor surface: %v", err)
	}
	verbose("created new wl_surface")

	layerSurface, err := interfaces.layerShell.GetLayerSurface(wlSurface,
		output.wlOutput, uint32(wlr_layer_shell.LayerShellLayerOverlay), "ephbar-overlay")
	if err != nil {
		fatal("unable to get layer_surface: %v", err)
	}
	verbose("got layer_surface")

	surface := &Surface{
		output:       output,
		wlSurface:    wlSurface,
		layerSurface: layerSurface,
	}
	output.surface = surface

	// Configure layerSurface
	layerSurface.AddConfigureHandler(surface.HandleLayerSurfaceConfigure)
	layerSurface.SetSize(0, height)
	layerSurface.SetAnchor(uint32(wlr_layer_shell.LayerSurfaceAnchorTop | wlr_layer_shell.LayerSurfaceAnchorLeft | wlr_layer_shell.LayerSurfaceAnchorRight))
	verbose("configured layer_surface")

	if err2 := wlSurface.Commit(); err2 != nil {
		fatal("unable to commit surface state: %v", err2)
	}
}

func (surface *Surface) HandleLayerSurfaceConfigure(e wlr_layer_shell.LayerSurfaceConfigureEvent) {
	// Send ack to xdg_surface that we have a frame.
	if err := surface.layerSurface.AckConfigure(e.Serial); err != nil {
		fatal("unable to ack layer surface configure")
	}

	surface.width = e.Width
	surface.height = e.Height

	// Draw frame
	buffer := surface.output.wayland.appState.drawFrame(surface)

	// Attach new frame to the surface
	if err := surface.wlSurface.Attach(buffer, 0, 0); err != nil {
		fatal("unable to attach buffer to surface: %v", err)
	}
	// Commit the surface state
	if err := surface.wlSurface.Commit(); err != nil {
		fatal("unable to commit surface state: %v", err)
	}
}

func (seat *Seat) HandleSeatCapabilities(e client.SeatCapabilitiesEvent) {
	hasPointer := (e.Capabilities * uint32(client.SeatCapabilityPointer)) != 0

	if hasPointer && seat.pointer == nil {
		seat.attachPointer()
	} else if !hasPointer && seat.pointer != nil {
		seat.releasePointer()
	}

	hasKeyboard := (e.Capabilities * uint32(client.SeatCapabilityKeyboard)) != 0

	if hasKeyboard && seat.keyboard == nil {
		seat.attachKeyboard()
	} else if !hasKeyboard && seat.keyboard != nil {
		seat.releaseKeyboard()
	}
}

func (wayland *WaylandState) Cleanup() {
	for _, seat := range wayland.seats {
		if seat.pointer != nil {
			seat.releasePointer()
		}
		if seat.keyboard != nil {
			seat.releaseKeyboard()
		}
	}

	for _, output := range wayland.outputs {
		if output.surface != nil {
			surface := output.surface

			if surface.layerSurface != nil {
				if err := surface.layerSurface.Destroy(); err != nil {
					info("unable to destroy layerSurface:", err)
				}
				surface.layerSurface = nil
			}

			if surface.wlSurface != nil {
				if err := surface.wlSurface.Destroy(); err != nil {
					info("unable to destroy wl_surface:", err)
				}
				surface.wlSurface = nil
			}
		}
	}

	interfaces := wayland.interfaces

	if interfaces.shm != nil {
		if err := interfaces.shm.Destroy(); err != nil {
			info("unable to destroy wl_shm:", err)
		}
		interfaces.shm = nil
	}

	if interfaces.compositor != nil {
		if err := interfaces.compositor.Destroy(); err != nil {
			info("unable to destroy wl_compositor:", err)
		}
		interfaces.compositor = nil
	}

	if wayland.registry != nil {
		if err := wayland.registry.Destroy(); err != nil {
			info("unable to destroy wl_registry:", err)
		}
		wayland.registry = nil
	}

	if wayland.display != nil {
		if err := wayland.display.Destroy(); err != nil {
			info("unable to destroy wl_display:", err)
		}
		wayland.display = nil
	}

	// Close the wayland server connection
	if err := wayland.context().Close(); err != nil {
		info("unable to close wayland context:", err)
	}
}
