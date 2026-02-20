package device

import (
	"fmt"
	"time"
)

const (
	reportID       = 0xEC
	reportLen      = 65
	controlTimeout = 1000 * time.Millisecond
	readTimeout    = 500 * time.Millisecond
)

// Per-bar color values from Windows capture
var colors = []byte{0xAA, 0xE6, 0xF0, 0xAA, 0xE6, 0xF0, 0xAA, 0xE6, 0xF0, 0xAA, 0xE6, 0xF0, 0xAA, 0xE6, 0xF0}

func (d *ProArtDevice) setReport(payload []byte) error {
	buf := make([]byte, reportLen)
	buf[0] = reportID
	copy(buf[1:], payload)

	// bmRequestType=0x21 (host-to-device, class, interface)
	// bRequest=0x09 (SET_REPORT)
	// wValue=0x02EC (Output report, report ID 0xEC)
	// wIndex=2 (HID interface number)
	_, err := d.usb.controlTransfer(0x21, 0x09, 0x0200|uint16(reportID), uint16(hidIfaceNum), buf, controlTimeout)
	return err
}

func (d *ProArtDevice) getReport() []byte {
	buf := make([]byte, reportLen)
	n, err := d.usb.interruptTransfer(epIn, buf, readTimeout)
	if err != nil {
		return nil
	}
	return buf[:n]
}

// Init runs the device initialization sequence (exact dump1 replay).
// Sequence: 4x(40 80 00 02 COLORS) -> B8 -> read -> 35 00 00 00 FF
func (d *ProArtDevice) Init() error {
	payload := append([]byte{0x40, 0x80, 0x00, 0x02}, colors...)
	for i := 0; i < 4; i++ {
		if err := d.setReport(payload); err != nil {
			return fmt.Errorf("init direct write %d: %w", i, err)
		}
	}

	if err := d.setReport([]byte{0xB8}); err != nil {
		return fmt.Errorf("init reset: %w", err)
	}

	d.getReport()

	if err := d.setReport([]byte{0x35, 0x00, 0x00, 0x00, 0xFF}); err != nil {
		return fmt.Errorf("init mode set: %w", err)
	}

	return nil
}

// SetMeterLevel sets the 5-bar LED meter to level (0-5).
// Sequence per dump1: B8 -> read -> 35 FF -> 5x(40 80 00 LEVEL COLORS)
func (d *ProArtDevice) SetMeterLevel(level int) error {
	if err := d.setReport([]byte{0xB8}); err != nil {
		return fmt.Errorf("meter reset: %w", err)
	}

	d.getReport()

	if err := d.setReport([]byte{0x35, 0x00, 0x00, 0x00, 0xFF}); err != nil {
		return fmt.Errorf("meter mode set: %w", err)
	}

	payload := append([]byte{0x40, 0x80, 0x00, byte(level)}, colors...)
	for i := 0; i < 5; i++ {
		if err := d.setReport(payload); err != nil {
			return fmt.Errorf("meter direct write %d: %w", i, err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	return nil
}
