package device

/*
#cgo pkg-config: libusb-1.0
#include <libusb.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"time"
	"unsafe"
)

type usbHandle struct {
	ctx    *C.libusb_context
	handle *C.libusb_device_handle
}

func usbOpen(vid, pid uint16) (*usbHandle, error) {
	var ctx *C.libusb_context
	if rc := C.libusb_init(&ctx); rc < 0 {
		return nil, fmt.Errorf("libusb_init: %s", C.GoString(C.libusb_strerror(C.int(rc))))
	}

	handle := C.libusb_open_device_with_vid_pid(ctx, C.uint16_t(vid), C.uint16_t(pid))
	if handle == nil {
		C.libusb_exit(ctx)
		return nil, fmt.Errorf("ASUS ProArt LC cooler not found (USB %04x:%04x)", vid, pid)
	}

	return &usbHandle{ctx: ctx, handle: handle}, nil
}

func (u *usbHandle) setAutoDetach() error {
	if rc := C.libusb_set_auto_detach_kernel_driver(u.handle, 1); rc < 0 {
		return fmt.Errorf("set auto detach: %s", C.GoString(C.libusb_strerror(C.int(rc))))
	}
	return nil
}

func (u *usbHandle) setConfiguration(cfgNum int) error {
	if rc := C.libusb_set_configuration(u.handle, C.int(cfgNum)); rc < 0 {
		return fmt.Errorf("set configuration: %s", C.GoString(C.libusb_strerror(C.int(rc))))
	}
	return nil
}

func (u *usbHandle) claimInterface(ifaceNum int) error {
	if rc := C.libusb_claim_interface(u.handle, C.int(ifaceNum)); rc < 0 {
		return fmt.Errorf("claim interface %d: %s", ifaceNum, C.GoString(C.libusb_strerror(C.int(rc))))
	}
	return nil
}

func (u *usbHandle) releaseInterface(ifaceNum int) {
	C.libusb_release_interface(u.handle, C.int(ifaceNum))
}

func (u *usbHandle) controlTransfer(bmRequestType uint8, bRequest uint8, wValue uint16, wIndex uint16, data []byte, timeout time.Duration) (int, error) {
	var dataPtr *C.uchar
	if len(data) > 0 {
		dataPtr = (*C.uchar)(unsafe.Pointer(&data[0]))
	}
	rc := C.libusb_control_transfer(
		u.handle,
		C.uint8_t(bmRequestType),
		C.uint8_t(bRequest),
		C.uint16_t(wValue),
		C.uint16_t(wIndex),
		dataPtr,
		C.uint16_t(len(data)),
		C.uint(timeout.Milliseconds()),
	)
	if rc < 0 {
		return 0, fmt.Errorf("control transfer: %s", C.GoString(C.libusb_strerror(C.int(rc))))
	}
	return int(rc), nil
}

func (u *usbHandle) interruptTransfer(endpoint uint8, data []byte, timeout time.Duration) (int, error) {
	var transferred C.int
	var dataPtr *C.uchar
	if len(data) > 0 {
		dataPtr = (*C.uchar)(unsafe.Pointer(&data[0]))
	}
	rc := C.libusb_interrupt_transfer(
		u.handle,
		C.uint8_t(endpoint),
		dataPtr,
		C.int(len(data)),
		&transferred,
		C.uint(timeout.Milliseconds()),
	)
	if rc < 0 {
		return 0, fmt.Errorf("interrupt transfer: %s", C.GoString(C.libusb_strerror(C.int(rc))))
	}
	return int(transferred), nil
}

func (u *usbHandle) close() {
	if u.handle != nil {
		C.libusb_close(u.handle)
	}
	if u.ctx != nil {
		C.libusb_exit(u.ctx)
	}
}
