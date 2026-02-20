package device

const (
	vid         = 0x0B05
	pid         = 0x1AFE
	hidIfaceNum = 2
	epIn        = 0x82
)

type ProArtDevice struct {
	usb *usbHandle
}

func Open() (*ProArtDevice, error) {
	usb, err := usbOpen(vid, pid)
	if err != nil {
		return nil, err
	}

	if err := usb.setAutoDetach(); err != nil {
		usb.close()
		return nil, err
	}

	if err := usb.setConfiguration(1); err != nil {
		usb.close()
		return nil, err
	}

	if err := usb.claimInterface(0); err != nil {
		usb.close()
		return nil, err
	}

	if err := usb.claimInterface(hidIfaceNum); err != nil {
		usb.releaseInterface(0)
		usb.close()
		return nil, err
	}

	return &ProArtDevice{usb: usb}, nil
}

func (d *ProArtDevice) Close() {
	d.usb.releaseInterface(hidIfaceNum)
	d.usb.releaseInterface(0)
	d.usb.close()
}
