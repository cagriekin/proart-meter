# proart-meter

Linux daemon that drives the front-panel LED temperature meter on ASUS ProArt LC 420 AIO liquid coolers. It reads the CPU temperature from the kernel hwmon subsystem and maps it to the cooler's 5-bar LED display over USB HID.

ASUS ships a Windows-only utility (ProArt Creator Hub) for this. There is no official Linux support. This project was built by reverse-engineering the USB protocol from packet captures of the Windows software.

## Hardware

| Field | Value |
|---|---|
| Device | ASUS ProArt LC 420 |
| USB Vendor ID | `0x0B05` (ASUSTeK) |
| USB Product ID | `0x1AFE` |
| USB Interface | 2 (HID) |
| Interrupt Endpoint | `0x82` (IN) |

The cooler exposes multiple USB interfaces. Interface 2 is the HID interface used to control the LED meter. The kernel's built-in HID driver must be detached before userspace can claim it (handled automatically via `libusb_set_auto_detach_kernel_driver`).

## USB Protocol

All communication uses HID SET_REPORT (Output) via USB control transfers, with an interrupt IN endpoint for device responses.

### Transport Layer

Every outbound message is a 65-byte HID Output report sent as a control transfer:

| Parameter | Value | Description |
|---|---|---|
| bmRequestType | `0x21` | Host-to-device, Class, Interface |
| bRequest | `0x09` | SET_REPORT |
| wValue | `0x02EC` | Report Type=Output (`0x02`), Report ID=`0xEC` |
| wIndex | `0x0002` | Interface 2 |
| wLength | `65` | Report ID byte + 64 bytes payload |

The first byte of the buffer is always the report ID (`0xEC`). The remaining 64 bytes carry the command payload, zero-padded.

Device responses are read from interrupt endpoint `0x82` as 65-byte transfers.

### Command Reference

#### Direct Write (`0x40`)

Sets bar colors and active level on the meter display.

```
Byte 0: 0x40  (command)
Byte 1: 0x80  (subcommand)
Byte 2: 0x00  (reserved)
Byte 3: level (0x00-0x05, number of bars to illuminate)
Byte 4+: RGB color data, 3 bytes per bar, 5 bars = 15 bytes
```

Color data is 5 triplets of `[R, G, B]`, one per bar from bottom to top.

#### Reset (`0xB8`)

Resets the meter controller state. Sent before every level change. The device responds with a status packet on the interrupt endpoint which must be read (and can be discarded).

```
Byte 0: 0xB8
```

#### Mode Set (`0x35`)

Configures the meter operating mode.

```
Byte 0: 0x35
Byte 1: 0x00
Byte 2: 0x00
Byte 3: 0x00
Byte 4: 0xFF
```

### Initialization Sequence

On startup, the following sequence is sent (replayed from Windows capture):

1. 4x Direct Write (`0x40 0x80 0x00 0x02 <colors>`) -- pre-load bars with color data
2. Reset (`0xB8`)
3. Read interrupt endpoint (discard response)
4. Mode Set (`0x35 0x00 0x00 0x00 0xFF`)

### Level Update Sequence

To change the meter level (0-5 bars illuminated):

1. Reset (`0xB8`)
2. Read interrupt endpoint (discard response)
3. Mode Set (`0x35 0x00 0x00 0x00 0xFF`)
4. 5x Direct Write (`0x40 0x80 0x00 <level> <colors>`) with 20ms delay between each

The level byte controls how many bars are lit. The 5 repeated writes with inter-packet delay are required -- the device does not reliably update with fewer.

## Temperature Mapping

The daemon reads CPU temperature from `/sys/class/hwmon/*/temp1_input` (matched by sensor name) and maps it to a meter level using 5 configurable thresholds:

```
temp < thresholds[0]           -> level 0 (all bars off)
temp >= thresholds[0]          -> level 1
temp >= thresholds[1]          -> level 2
temp >= thresholds[2]          -> level 3
temp >= thresholds[3]          -> level 4
temp >= thresholds[4]          -> level 5 (all bars on)
```

The meter only updates on level changes, not every poll cycle.

## Configuration

`/etc/proart-meter/config.yaml`:

```yaml
sensor: "k10temp"
poll_interval_seconds: 2
thresholds: [40, 50, 60, 70, 80]
```

| Field | Type | Description |
|---|---|---|
| `sensor` | string | hwmon sensor name (e.g. `k10temp` for AMD, `coretemp` for Intel) |
| `poll_interval_seconds` | int | Seconds between temperature reads |
| `thresholds` | int[5] | Temperature thresholds in Celsius, ascending order, exactly 5 values |

Find your sensor name:

```sh
cat /sys/class/hwmon/hwmon*/name
```

## Dependencies

- `libusb` -- USB device communication
- `lm_sensors` -- kernel modules for hardware temperature monitoring

## Building from Source

```sh
git clone https://github.com/cagriekin/proart-meter.git
cd proart-meter
make build
sudo make install
```

## Arch Linux (AUR)

```sh
yay -S proart-meter-git
```

Or build the package manually:

```sh
git clone https://github.com/cagriekin/proart-meter.git
cd proart-meter
makepkg -si
```

## Usage

```sh
# Start the service
sudo systemctl enable --now proart-meter

# Check status
sudo systemctl status proart-meter

# View logs
journalctl -u proart-meter -f
```

The daemon requires root or appropriate USB permissions to access the device. On shutdown (SIGINT/SIGTERM), it resets the meter to level 0 before exiting.

## License

MIT
