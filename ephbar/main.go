package main

import (
	"image"

	"github.com/fogleman/gg"

	"github.com/rajveermalviya/go-wayland/wayland/client"
	wlr_layer_shell "github.com/rajveermalviya/go-wayland/wayland/unstable/wlr-layer-shell-v1"

	memfd "github.com/justincormack/go-memfd"
	"golang.org/x/sys/unix"
)

// Global app state
type appState struct {
	width, height int32
	exit          bool

	// Registry interfaces
	registry   *client.Registry
	display    *client.Display
	output     *client.Output
	shm        *client.Shm
	compositor *client.Compositor
	seat       *client.Seat
	layerShell *wlr_layer_shell.LayerShell

	// Surfaces
	wlSurface    *client.Surface
	layerSurface *wlr_layer_shell.LayerSurface

	// Seats
	keyboard *client.Keyboard
	pointer  *client.Pointer

	// TODO: Create pointerState/keyboardState instead of storing directly in appState
	pointerEvent pointerEvent
}

func main() {
	app := &appState{
		width:  300,
		height: 300,
	}

	info("initializing")
	app.initWindow()

	// Start the dispatch loop
	info("starting")
	for !app.exit {
		app.dispatch()
	}

	info("closing")
	app.cleanup()
}

func (app *appState) initWindow() {
	// Connect to wayland server
	display, err := client.Connect("")
	if err != nil {
		fatal("unable to connect to wayland server: %v", err)
	}
	app.display = display

	// Get global interfaces registry
	registry, err := app.display.GetRegistry()
	if err != nil {
		fatal("unable to get global registry object: %v", err)
	}
	app.registry = registry

	// Add global interfaces registrar handler
	registry.AddGlobalHandler(app.HandleRegistryGlobal)
	// Wait for interfaces to register
	app.displayRoundTrip()
	// Wait for handler events
	app.displayRoundTrip()

	verbose("all interfaces registered")

	// Create a wl_surface for toplevel window
	wl_surface, err := app.compositor.CreateSurface()
	if err != nil {
		fatal("unable to create compositor surface: %v", err)
	}
	app.wlSurface = wl_surface
	verbose("created new wl_surface")

	layerSurface, err := app.layerShell.GetLayerSurface(wl_surface,
		app.output, uint32(wlr_layer_shell.LayerShellLayerOverlay), "ephbar-overlay")
	if err != nil {
		fatal("unable to get layer_surface: %v", err)
	}
	app.layerSurface = layerSurface
	verbose("got layer_surface")

	// Configure layerSurface
	layerSurface.AddConfigureHandler(app.HandleLayerSurfaceConfigure)
	layerSurface.SetSize(uint32(app.width), uint32(app.height))
	layerSurface.SetAnchor(uint32(wlr_layer_shell.LayerSurfaceAnchorTop))
	verbose("configured layer_surface")

	if err2 := app.wlSurface.Commit(); err2 != nil {
		fatal("unable to commit surface state: %v", err2)
	}
}

func (app *appState) dispatch() {
	app.display.Context().Dispatch()
}

func (app *appState) context() *client.Context {
	return app.display.Context()
}

func (app *appState) HandleRegistryGlobal(e client.RegistryGlobalEvent) {
	verbose("discovered an interface: %q", e.Interface)

	switch e.Interface {
	case "wl_compositor":
		compositor := client.NewCompositor(app.context())
		err := app.registry.Bind(e.Name, e.Interface, e.Version, compositor)
		if err != nil {
			fatal("unable to bind wl_compositor interface: %v", err)
		}
		app.compositor = compositor
	case "wl_shm":
		shm := client.NewShm(app.context())
		err := app.registry.Bind(e.Name, e.Interface, e.Version, shm)
		if err != nil {
			fatal("unable to bind wl_shm interface: %v", err)
		}
		app.shm = shm
	case "wl_seat":
		seat := client.NewSeat(app.context())
		err := app.registry.Bind(e.Name, e.Interface, e.Version, seat)
		if err != nil {
			fatal("unable to bind wl_seat interface: %v", err)
		}
		app.seat = seat
		seat.AddCapabilitiesHandler(app.HandleSeatCapabilities)
	case "wl_output":
		// TODO: Elegantly handle multiple outputs
		if app.output != nil {
			return
		}
		output := client.NewOutput(app.context())
		err := app.registry.Bind(e.Name, e.Interface, e.Version, output)
		if err != nil {
			fatal("unable to bind wl_output interface: %v", err)
		}
		app.output = output

	case "zwlr_layer_shell_v1":
		layer_shell := wlr_layer_shell.NewLayerShell(app.context())
		err := app.registry.Bind(e.Name, e.Interface, e.Version, layer_shell)
		if err != nil {
			fatal("unable to bind wlr_layer_shell interface: %v", err)
		}
		app.layerShell = layer_shell
	}
}

func (app *appState) HandleLayerSurfaceConfigure(e wlr_layer_shell.LayerSurfaceConfigureEvent) {
	// Send ack to xdg_surface that we have a frame.
	if err := app.layerSurface.AckConfigure(e.Serial); err != nil {
		fatal("unable to ack layer surface configure")
	}

	// Draw frame
	buffer := app.drawFrame()

	// Attach new frame to the surface
	if err := app.wlSurface.Attach(buffer, 0, 0); err != nil {
		fatal("unable to attach buffer to surface: %v", err)
	}
	// Commit the surface state
	if err := app.wlSurface.Commit(); err != nil {
		fatal("unable to commit surface state: %v", err)
	}
}

func (app *appState) drawFrame() *client.Buffer {
	// Create in-memory file descriptor
	fd, err := memfd.Create()
	if err != nil {
		fatal("unable to create file descriptor: %v", err)
	}
	defer fd.Close()

	// Resize fd
	stride := app.width * 4
	size := stride * app.height
	unix.Ftruncate(int(fd.Fd()), int64(size))

	// Map memory
	data, err := unix.Mmap(int(fd.Fd()), 0, int(size), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		fatal("unable to create mapping: %v", err)
	}
	defer unix.Munmap(data)

	// Created shm pool
	pool, err := app.shm.CreatePool(fd.Fd(), size)
	if err != nil {
		fatal("unable to create shm pool: %v", err)
	}
	defer pool.Destroy()

	// Create client buffer
	buf, err := pool.CreateBuffer(0, app.width, app.height, stride, uint32(client.ShmFormatArgb8888))
	if err != nil {
		fatal("unable to create client.Buffer from shm pool: %v", err)
	}

	// Draw
	img := image.NewRGBA(image.Rect(0, 0, int(app.width), int(app.height)))
	img.Pix = data

	dc := gg.NewContextForRGBA(img)
	dc.DrawRectangle(0, 0, float64(app.width), float64(app.height))
	dc.SetRGB(0, 0, 0)
	dc.Fill()
	dc.DrawCircle(150, 150, 50)
	dc.SetRGB(0, 0, 1)
	dc.Fill()

	// Destroy buffer on release
	buf.AddReleaseHandler(func(_ client.BufferReleaseEvent) {
		if err := buf.Destroy(); err != nil {
			fatal("unable to destroy buffer: %v", err)
		}
	})

	// Return buffer
	return buf
}

func (app *appState) HandleSeatCapabilities(e client.SeatCapabilitiesEvent) {
	havePointer := (e.Capabilities * uint32(client.SeatCapabilityPointer)) != 0

	if havePointer && app.pointer == nil {
		app.attachPointer()
	} else if !havePointer && app.pointer != nil {
		app.releasePointer()
	}

	haveKeyboard := (e.Capabilities * uint32(client.SeatCapabilityKeyboard)) != 0

	if haveKeyboard && app.keyboard == nil {
		app.attachKeyboard()
	} else if !haveKeyboard && app.keyboard != nil {
		app.releaseKeyboard()
	}
}

func (app *appState) displayRoundTrip() {
	// Get display sync callback
	callback, err := app.display.Sync()
	if err != nil {
		fatal("unable to get sync callback: %v", err)
	}
	defer callback.Destroy()

	done := false
	callback.AddDoneHandler(func(_ client.CallbackDoneEvent) {
		done = true
	})

	// Wait for callback to return
	for !done {
		app.dispatch()
	}
}

func (app *appState) cleanup() {
	// Release the pointer if registered
	if app.pointer != nil {
		app.releasePointer()
	}

	// Release the keyboard if registered
	if app.keyboard != nil {
		app.releaseKeyboard()
	}

	if app.layerSurface != nil {
		if err := app.layerSurface.Destroy(); err != nil {
			info("unable to destroy layerSurface:", err)
		}
		app.layerSurface = nil
	}

	if app.wlSurface != nil {
		if err := app.wlSurface.Destroy(); err != nil {
			info("unable to destroy wl_surface:", err)
		}
		app.wlSurface = nil
	}

	// Release wl_seat handlers
	if app.seat != nil {
		if err := app.seat.Release(); err != nil {
			info("unable to destroy wl_seat:", err)
		}
		app.seat = nil
	}

	if app.shm != nil {
		if err := app.shm.Destroy(); err != nil {
			info("unable to destroy wl_shm:", err)
		}
		app.shm = nil
	}

	if app.compositor != nil {
		if err := app.compositor.Destroy(); err != nil {
			info("unable to destroy wl_compositor:", err)
		}
		app.compositor = nil
	}

	if app.registry != nil {
		if err := app.registry.Destroy(); err != nil {
			info("unable to destroy wl_registry:", err)
		}
		app.registry = nil
	}

	if app.display != nil {
		if err := app.display.Destroy(); err != nil {
			info("unable to destroy wl_display:", err)
		}
	}

	// Close the wayland server connection
	if err := app.context().Close(); err != nil {
		info("unable to close wayland context:", err)
	}
}
