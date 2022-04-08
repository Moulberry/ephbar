package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/danielgatis/go-findfont/findfont"
	"github.com/fogleman/gg"

	"github.com/rajveermalviya/go-wayland/wayland/client"

	memfd "github.com/justincormack/go-memfd"
	"golang.org/x/sys/unix"
)

// Global app state
type AppState struct {
	exit bool

	wayland *WaylandState

	// TODO: Create pointerState/keyboardState instead of storing directly in AppState
	pointerEvent pointerEvent
}

func main() {
	app := &AppState{}

	info("initializing")
	app.initWindow()

	// Start the dispatch loop
	info("starting")
	for !app.exit {
		app.wayland.dispatch()
	}

	info("closing")
	app.cleanup()
}

func (app *AppState) initWindow() {
	app.wayland = NewWaylandState(app)

	app.wayland.RegisterGlobals()
}

func makeColor(col uint32) *color.RGBA {
	return &color.RGBA{
		uint8(col >> 8),
		uint8(col >> 16),
		uint8(col >> 24),
		uint8(col),
	}
}

func (app *AppState) drawFrame(surface *Surface) *client.Buffer {
	// Create in-memory file descriptor
	fd, err := memfd.Create()
	if err != nil {
		fatal("unable to create file descriptor: %v", err)
	}
	defer fd.Close()

	width := surface.width
	height := surface.height

	// Resize fd
	stride := int32(width * 4)
	size := stride * int32(height)
	unix.Ftruncate(int(fd.Fd()), int64(size))

	// Map memory
	data, err := unix.Mmap(int(fd.Fd()), 0, int(size), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		fatal("unable to create mapping: %v", err)
	}
	defer unix.Munmap(data)

	// Created shm pool
	pool, err := app.wayland.interfaces.shm.CreatePool(fd.Fd(), size)
	if err != nil {
		fatal("unable to create shm pool: %v", err)
	}
	defer pool.Destroy()

	// Create client buffer
	buf, err := pool.CreateBuffer(0, int32(width), int32(height), stride, uint32(client.ShmFormatArgb8888))
	if err != nil {
		fatal("unable to create client.Buffer from shm pool: %v", err)
	}

	// Draw
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	img.Pix = data

	fonts, err := findfont.Find("monospace:pixelsize=14", findfont.FontBold)
	if err != nil {
		fatal("Couldn't find monospace font on system: %v", err)
	}
	font := fonts[0][2]
	info("Using font: %s", font)

	dc := gg.NewContextForRGBA(img)

	//color_base := &color.RGBA{0x22, 0x22, 0x22, 0xFF}
	//color_bg_inactive := &color.RGBA{0x66, 0x55, 0x55, 0xFF}
	color_bg_active := makeColor(0x005577FF)
	//color_fg_active := &color.RGBA{0xb3, 0xb3, 0xb3, 0xFF}
	//color_fg_inactive := &color.RGBA{0x70, 0x70, 0x70, 0xFF}

	color_base := makeColor(0x2E3440FF)
	color_bg_inactive := makeColor(0x434c5eFF)
	color_fg_active := makeColor(0xb3b3b3ff) //makeColor(0x81a1c1FF)

	dc.DrawRoundedRectangle(8, 8, 9*30+4, 25+4, 5)
	dc.SetColor(color_bg_inactive)
	dc.Fill()
	dc.DrawRoundedRectangle(10, 10, 9*30, 25, 5)
	dc.SetColor(color_base)
	dc.Fill()

	for i := 0; i < 9; i++ {
		x := float64(10 + i*30)
		y := float64(10)

		if i == 3 {
			dc.DrawRoundedRectangle(x-1, y, 27, 25, 5)
			dc.SetColor(color_bg_active)
			dc.Fill()
		}

		dc.SetColor(color_fg_active)
		dc.LoadFontFace(font, 20)
		dc.DrawString(fmt.Sprintf("%d", i+1), x+6, y+20)
		dc.Fill()
	}

	// Destroy buffer on release
	buf.AddReleaseHandler(func(_ client.BufferReleaseEvent) {
		if err := buf.Destroy(); err != nil {
			fatal("unable to destroy buffer: %v", err)
		}
	})

	// Return buffer
	return buf
}

func (app *AppState) cleanup() {
	app.wayland.Cleanup()
}
