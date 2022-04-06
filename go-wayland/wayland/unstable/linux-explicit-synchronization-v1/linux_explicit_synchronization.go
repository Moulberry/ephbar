// Generated by go-wayland-scanner
// https://github.com/rajveermalviya/go-wayland/cmd/go-wayland-scanner
// XML file : https://raw.githubusercontent.com/wayland-project/wayland-protocols/1.25/unstable/linux-explicit-synchronization/linux-explicit-synchronization-unstable-v1.xml
//
// zwp_linux_explicit_synchronization_unstable_v1 Protocol Copyright:
//
// Copyright 2016 The Chromium Authors.
// Copyright 2017 Intel Corporation
// Copyright 2018 Collabora, Ltd
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice (including the next
// paragraph) shall be included in all copies or substantial portions of the
// Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package linux_explicit_synchronization

import (
	"github.com/rajveermalviya/go-wayland/wayland/client"
	"golang.org/x/sys/unix"
)

// LinuxExplicitSynchronization : protocol for providing explicit synchronization
//
// This global is a factory interface, allowing clients to request
// explicit synchronization for buffers on a per-surface basis.
//
// See zwp_linux_surface_synchronization_v1 for more information.
//
// This interface is derived from Chromium's
// zcr_linux_explicit_synchronization_v1.
//
// Warning! The protocol described in this file is experimental and
// backward incompatible changes may be made. Backward compatible changes
// may be added together with the corresponding interface version bump.
// Backward incompatible changes are done by bumping the version number in
// the protocol and interface names and resetting the interface version.
// Once the protocol is to be declared stable, the 'z' prefix and the
// version number in the protocol and interface names are removed and the
// interface version number is reset.
type LinuxExplicitSynchronization struct {
	client.BaseProxy
}

// NewLinuxExplicitSynchronization : protocol for providing explicit synchronization
//
// This global is a factory interface, allowing clients to request
// explicit synchronization for buffers on a per-surface basis.
//
// See zwp_linux_surface_synchronization_v1 for more information.
//
// This interface is derived from Chromium's
// zcr_linux_explicit_synchronization_v1.
//
// Warning! The protocol described in this file is experimental and
// backward incompatible changes may be made. Backward compatible changes
// may be added together with the corresponding interface version bump.
// Backward incompatible changes are done by bumping the version number in
// the protocol and interface names and resetting the interface version.
// Once the protocol is to be declared stable, the 'z' prefix and the
// version number in the protocol and interface names are removed and the
// interface version number is reset.
func NewLinuxExplicitSynchronization(ctx *client.Context) *LinuxExplicitSynchronization {
	zwpLinuxExplicitSynchronizationV1 := &LinuxExplicitSynchronization{}
	ctx.Register(zwpLinuxExplicitSynchronizationV1)
	return zwpLinuxExplicitSynchronizationV1
}

// Destroy : destroy explicit synchronization factory object
//
// Destroy this explicit synchronization factory object. Other objects,
// including zwp_linux_surface_synchronization_v1 objects created by this
// factory, shall not be affected by this request.
//
func (i *LinuxExplicitSynchronization) Destroy() error {
	defer i.Context().Unregister(i)
	const opcode = 0
	const rLen = 8
	r := make([]byte, rLen)
	l := 0
	client.PutUint32(r[l:4], i.ID())
	l += 4
	client.PutUint32(r[l:l+4], uint32(rLen<<16|opcode&0x0000ffff))
	l += 4
	err := i.Context().WriteMsg(r, nil)
	return err
}

// GetSynchronization : extend surface interface for explicit synchronization
//
// Instantiate an interface extension for the given wl_surface to provide
// explicit synchronization.
//
// If the given wl_surface already has an explicit synchronization object
// associated, the synchronization_exists protocol error is raised.
//
// Graphics APIs, like EGL or Vulkan, that manage the buffer queue and
// commits of a wl_surface themselves, are likely to be using this
// extension internally. If a client is using such an API for a
// wl_surface, it should not directly use this extension on that surface,
// to avoid raising a synchronization_exists protocol error.
//
//  surface: the surface
func (i *LinuxExplicitSynchronization) GetSynchronization(surface *client.Surface) (*LinuxSurfaceSynchronization, error) {
	id := NewLinuxSurfaceSynchronization(i.Context())
	const opcode = 1
	const rLen = 8 + 4 + 4
	r := make([]byte, rLen)
	l := 0
	client.PutUint32(r[l:4], i.ID())
	l += 4
	client.PutUint32(r[l:l+4], uint32(rLen<<16|opcode&0x0000ffff))
	l += 4
	client.PutUint32(r[l:l+4], id.ID())
	l += 4
	client.PutUint32(r[l:l+4], surface.ID())
	l += 4
	err := i.Context().WriteMsg(r, nil)
	return id, err
}

type LinuxExplicitSynchronizationError uint32

// LinuxExplicitSynchronizationError :
const (
	// LinuxExplicitSynchronizationErrorSynchronizationExists : the surface already has a synchronization object associated
	LinuxExplicitSynchronizationErrorSynchronizationExists LinuxExplicitSynchronizationError = 0
)

func (e LinuxExplicitSynchronizationError) Name() string {
	switch e {
	case LinuxExplicitSynchronizationErrorSynchronizationExists:
		return "synchronization_exists"
	default:
		return ""
	}
}

func (e LinuxExplicitSynchronizationError) Value() string {
	switch e {
	case LinuxExplicitSynchronizationErrorSynchronizationExists:
		return "0"
	default:
		return ""
	}
}

func (e LinuxExplicitSynchronizationError) String() string {
	return e.Name() + "=" + e.Value()
}

// LinuxSurfaceSynchronization : per-surface explicit synchronization support
//
// This object implements per-surface explicit synchronization.
//
// Synchronization refers to co-ordination of pipelined operations performed
// on buffers. Most GPU clients will schedule an asynchronous operation to
// render to the buffer, then immediately send the buffer to the compositor
// to be attached to a surface.
//
// In implicit synchronization, ensuring that the rendering operation is
// complete before the compositor displays the buffer is an implementation
// detail handled by either the kernel or userspace graphics driver.
//
// By contrast, in explicit synchronization, dma_fence objects mark when the
// asynchronous operations are complete. When submitting a buffer, the
// client provides an acquire fence which will be waited on before the
// compositor accesses the buffer. The Wayland server, through a
// zwp_linux_buffer_release_v1 object, will inform the client with an event
// which may be accompanied by a release fence, when the compositor will no
// longer access the buffer contents due to the specific commit that
// requested the release event.
//
// Each surface can be associated with only one object of this interface at
// any time.
//
// In version 1 of this interface, explicit synchronization is only
// guaranteed to be supported for buffers created with any version of the
// wp_linux_dmabuf buffer factory. Version 2 additionally guarantees
// explicit synchronization support for opaque EGL buffers, which is a type
// of platform specific buffers described in the EGL_WL_bind_wayland_display
// extension. Compositors are free to support explicit synchronization for
// additional buffer types.
type LinuxSurfaceSynchronization struct {
	client.BaseProxy
}

// NewLinuxSurfaceSynchronization : per-surface explicit synchronization support
//
// This object implements per-surface explicit synchronization.
//
// Synchronization refers to co-ordination of pipelined operations performed
// on buffers. Most GPU clients will schedule an asynchronous operation to
// render to the buffer, then immediately send the buffer to the compositor
// to be attached to a surface.
//
// In implicit synchronization, ensuring that the rendering operation is
// complete before the compositor displays the buffer is an implementation
// detail handled by either the kernel or userspace graphics driver.
//
// By contrast, in explicit synchronization, dma_fence objects mark when the
// asynchronous operations are complete. When submitting a buffer, the
// client provides an acquire fence which will be waited on before the
// compositor accesses the buffer. The Wayland server, through a
// zwp_linux_buffer_release_v1 object, will inform the client with an event
// which may be accompanied by a release fence, when the compositor will no
// longer access the buffer contents due to the specific commit that
// requested the release event.
//
// Each surface can be associated with only one object of this interface at
// any time.
//
// In version 1 of this interface, explicit synchronization is only
// guaranteed to be supported for buffers created with any version of the
// wp_linux_dmabuf buffer factory. Version 2 additionally guarantees
// explicit synchronization support for opaque EGL buffers, which is a type
// of platform specific buffers described in the EGL_WL_bind_wayland_display
// extension. Compositors are free to support explicit synchronization for
// additional buffer types.
func NewLinuxSurfaceSynchronization(ctx *client.Context) *LinuxSurfaceSynchronization {
	zwpLinuxSurfaceSynchronizationV1 := &LinuxSurfaceSynchronization{}
	ctx.Register(zwpLinuxSurfaceSynchronizationV1)
	return zwpLinuxSurfaceSynchronizationV1
}

// Destroy : destroy synchronization object
//
// Destroy this explicit synchronization object.
//
// Any fence set by this object with set_acquire_fence since the last
// commit will be discarded by the server. Any fences set by this object
// before the last commit are not affected.
//
// zwp_linux_buffer_release_v1 objects created by this object are not
// affected by this request.
//
func (i *LinuxSurfaceSynchronization) Destroy() error {
	defer i.Context().Unregister(i)
	const opcode = 0
	const rLen = 8
	r := make([]byte, rLen)
	l := 0
	client.PutUint32(r[l:4], i.ID())
	l += 4
	client.PutUint32(r[l:l+4], uint32(rLen<<16|opcode&0x0000ffff))
	l += 4
	err := i.Context().WriteMsg(r, nil)
	return err
}

// SetAcquireFence : set the acquire fence
//
// Set the acquire fence that must be signaled before the compositor
// may sample from the buffer attached with wl_surface.attach. The fence
// is a dma_fence kernel object.
//
// The acquire fence is double-buffered state, and will be applied on the
// next wl_surface.commit request for the associated surface. Thus, it
// applies only to the buffer that is attached to the surface at commit
// time.
//
// If the provided fd is not a valid dma_fence fd, then an INVALID_FENCE
// error is raised.
//
// If a fence has already been attached during the same commit cycle, a
// DUPLICATE_FENCE error is raised.
//
// If the associated wl_surface was destroyed, a NO_SURFACE error is
// raised.
//
// If at surface commit time the attached buffer does not support explicit
// synchronization, an UNSUPPORTED_BUFFER error is raised.
//
// If at surface commit time there is no buffer attached, a NO_BUFFER
// error is raised.
//
//  fd: acquire fence fd
func (i *LinuxSurfaceSynchronization) SetAcquireFence(fd uintptr) error {
	const opcode = 1
	const rLen = 8
	r := make([]byte, rLen)
	l := 0
	client.PutUint32(r[l:4], i.ID())
	l += 4
	client.PutUint32(r[l:l+4], uint32(rLen<<16|opcode&0x0000ffff))
	l += 4
	oob := unix.UnixRights(int(fd))
	err := i.Context().WriteMsg(r, oob)
	return err
}

// GetRelease : release fence for last-attached buffer
//
// Create a listener for the release of the buffer attached by the
// client with wl_surface.attach. See zwp_linux_buffer_release_v1
// documentation for more information.
//
// The release object is double-buffered state, and will be associated
// with the buffer that is attached to the surface at wl_surface.commit
// time.
//
// If a zwp_linux_buffer_release_v1 object has already been requested for
// the surface in the same commit cycle, a DUPLICATE_RELEASE error is
// raised.
//
// If the associated wl_surface was destroyed, a NO_SURFACE error
// is raised.
//
// If at surface commit time there is no buffer attached, a NO_BUFFER
// error is raised.
//
func (i *LinuxSurfaceSynchronization) GetRelease() (*LinuxBufferRelease, error) {
	release := NewLinuxBufferRelease(i.Context())
	const opcode = 2
	const rLen = 8 + 4
	r := make([]byte, rLen)
	l := 0
	client.PutUint32(r[l:4], i.ID())
	l += 4
	client.PutUint32(r[l:l+4], uint32(rLen<<16|opcode&0x0000ffff))
	l += 4
	client.PutUint32(r[l:l+4], release.ID())
	l += 4
	err := i.Context().WriteMsg(r, nil)
	return release, err
}

type LinuxSurfaceSynchronizationError uint32

// LinuxSurfaceSynchronizationError :
const (
	// LinuxSurfaceSynchronizationErrorInvalidFence : the fence specified by the client could not be imported
	LinuxSurfaceSynchronizationErrorInvalidFence LinuxSurfaceSynchronizationError = 0
	// LinuxSurfaceSynchronizationErrorDuplicateFence : multiple fences added for a single surface commit
	LinuxSurfaceSynchronizationErrorDuplicateFence LinuxSurfaceSynchronizationError = 1
	// LinuxSurfaceSynchronizationErrorDuplicateRelease : multiple releases added for a single surface commit
	LinuxSurfaceSynchronizationErrorDuplicateRelease LinuxSurfaceSynchronizationError = 2
	// LinuxSurfaceSynchronizationErrorNoSurface : the associated wl_surface was destroyed
	LinuxSurfaceSynchronizationErrorNoSurface LinuxSurfaceSynchronizationError = 3
	// LinuxSurfaceSynchronizationErrorUnsupportedBuffer : the buffer does not support explicit synchronization
	LinuxSurfaceSynchronizationErrorUnsupportedBuffer LinuxSurfaceSynchronizationError = 4
	// LinuxSurfaceSynchronizationErrorNoBuffer : no buffer was attached
	LinuxSurfaceSynchronizationErrorNoBuffer LinuxSurfaceSynchronizationError = 5
)

func (e LinuxSurfaceSynchronizationError) Name() string {
	switch e {
	case LinuxSurfaceSynchronizationErrorInvalidFence:
		return "invalid_fence"
	case LinuxSurfaceSynchronizationErrorDuplicateFence:
		return "duplicate_fence"
	case LinuxSurfaceSynchronizationErrorDuplicateRelease:
		return "duplicate_release"
	case LinuxSurfaceSynchronizationErrorNoSurface:
		return "no_surface"
	case LinuxSurfaceSynchronizationErrorUnsupportedBuffer:
		return "unsupported_buffer"
	case LinuxSurfaceSynchronizationErrorNoBuffer:
		return "no_buffer"
	default:
		return ""
	}
}

func (e LinuxSurfaceSynchronizationError) Value() string {
	switch e {
	case LinuxSurfaceSynchronizationErrorInvalidFence:
		return "0"
	case LinuxSurfaceSynchronizationErrorDuplicateFence:
		return "1"
	case LinuxSurfaceSynchronizationErrorDuplicateRelease:
		return "2"
	case LinuxSurfaceSynchronizationErrorNoSurface:
		return "3"
	case LinuxSurfaceSynchronizationErrorUnsupportedBuffer:
		return "4"
	case LinuxSurfaceSynchronizationErrorNoBuffer:
		return "5"
	default:
		return ""
	}
}

func (e LinuxSurfaceSynchronizationError) String() string {
	return e.Name() + "=" + e.Value()
}

// LinuxBufferRelease : buffer release explicit synchronization
//
// This object is instantiated in response to a
// zwp_linux_surface_synchronization_v1.get_release request.
//
// It provides an alternative to wl_buffer.release events, providing a
// unique release from a single wl_surface.commit request. The release event
// also supports explicit synchronization, providing a fence FD for the
// client to synchronize against.
//
// Exactly one event, either a fenced_release or an immediate_release, will
// be emitted for the wl_surface.commit request. The compositor can choose
// release by release which event it uses.
//
// This event does not replace wl_buffer.release events; servers are still
// required to send those events.
//
// Once a buffer release object has delivered a 'fenced_release' or an
// 'immediate_release' event it is automatically destroyed.
type LinuxBufferRelease struct {
	client.BaseProxy
	fencedReleaseHandlers    []LinuxBufferReleaseFencedReleaseHandlerFunc
	immediateReleaseHandlers []LinuxBufferReleaseImmediateReleaseHandlerFunc
}

// NewLinuxBufferRelease : buffer release explicit synchronization
//
// This object is instantiated in response to a
// zwp_linux_surface_synchronization_v1.get_release request.
//
// It provides an alternative to wl_buffer.release events, providing a
// unique release from a single wl_surface.commit request. The release event
// also supports explicit synchronization, providing a fence FD for the
// client to synchronize against.
//
// Exactly one event, either a fenced_release or an immediate_release, will
// be emitted for the wl_surface.commit request. The compositor can choose
// release by release which event it uses.
//
// This event does not replace wl_buffer.release events; servers are still
// required to send those events.
//
// Once a buffer release object has delivered a 'fenced_release' or an
// 'immediate_release' event it is automatically destroyed.
func NewLinuxBufferRelease(ctx *client.Context) *LinuxBufferRelease {
	zwpLinuxBufferReleaseV1 := &LinuxBufferRelease{}
	ctx.Register(zwpLinuxBufferReleaseV1)
	return zwpLinuxBufferReleaseV1
}

func (i *LinuxBufferRelease) Destroy() error {
	i.Context().Unregister(i)
	return nil
}

// LinuxBufferReleaseFencedReleaseEvent : release buffer with fence
//
// Sent when the compositor has finalised its usage of the associated
// buffer for the relevant commit, providing a dma_fence which will be
// signaled when all operations by the compositor on that buffer for that
// commit have finished.
//
// Once the fence has signaled, and assuming the associated buffer is not
// pending release from other wl_surface.commit requests, no additional
// explicit or implicit synchronization is required to safely reuse or
// destroy the buffer.
//
// This event destroys the zwp_linux_buffer_release_v1 object.
type LinuxBufferReleaseFencedReleaseEvent struct {
	Fence uintptr
}
type LinuxBufferReleaseFencedReleaseHandlerFunc func(LinuxBufferReleaseFencedReleaseEvent)

// AddFencedReleaseHandler : adds handler for LinuxBufferReleaseFencedReleaseEvent
func (i *LinuxBufferRelease) AddFencedReleaseHandler(f LinuxBufferReleaseFencedReleaseHandlerFunc) {
	if f == nil {
		return
	}

	i.fencedReleaseHandlers = append(i.fencedReleaseHandlers, f)
}

// LinuxBufferReleaseImmediateReleaseEvent : release buffer immediately
//
// Sent when the compositor has finalised its usage of the associated
// buffer for the relevant commit, and either performed no operations
// using it, or has a guarantee that all its operations on that buffer for
// that commit have finished.
//
// Once this event is received, and assuming the associated buffer is not
// pending release from other wl_surface.commit requests, no additional
// explicit or implicit synchronization is required to safely reuse or
// destroy the buffer.
//
// This event destroys the zwp_linux_buffer_release_v1 object.
type LinuxBufferReleaseImmediateReleaseEvent struct{}
type LinuxBufferReleaseImmediateReleaseHandlerFunc func(LinuxBufferReleaseImmediateReleaseEvent)

// AddImmediateReleaseHandler : adds handler for LinuxBufferReleaseImmediateReleaseEvent
func (i *LinuxBufferRelease) AddImmediateReleaseHandler(f LinuxBufferReleaseImmediateReleaseHandlerFunc) {
	if f == nil {
		return
	}

	i.immediateReleaseHandlers = append(i.immediateReleaseHandlers, f)
}

func (i *LinuxBufferRelease) Dispatch(opcode uint16, fd uintptr, data []byte) {
	switch opcode {
	case 0:
		if len(i.fencedReleaseHandlers) == 0 {
			return
		}
		var e LinuxBufferReleaseFencedReleaseEvent
		e.Fence = fd
		for _, f := range i.fencedReleaseHandlers {
			f(e)
		}
	case 1:
		if len(i.immediateReleaseHandlers) == 0 {
			return
		}
		var e LinuxBufferReleaseImmediateReleaseEvent
		for _, f := range i.immediateReleaseHandlers {
			f(e)
		}
	}
}
