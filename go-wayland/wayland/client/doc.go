// Package client is Go port of wayland-client library
// for writing pure Go GUI software for wayland supported
// platforms.
package client

//go:generate go-wayland-scanner -pkg client -prefix wl -o client.go -i https://raw.githubusercontent.com/wayland-project/wayland/1.20.0/protocol/wayland.xml
