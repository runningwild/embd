// Support for the very popular DS18B20 1-WIre temperature sensor
package ds18b20
import (
	"github.com/zlowred/embd"
	"errors"
	"sync"
)

type DS18B20_Resolution int

const (
	Resolution_9bit DS18B20_Resolution = iota
	Resolution_10bit
	Resolution_11bit
	Resolution_12bit
)

var ds18b20_crc_data = [...]byte{
	0, 94, 188, 226, 97, 63, 221, 131, 194, 156, 126, 32, 163, 253, 31, 65,
	157, 195, 33, 127, 252, 162, 64, 30, 95, 1, 227, 189, 62, 96, 130, 220,
	35, 125, 159, 193, 66, 28, 254, 160, 225, 191, 93, 3, 128, 222, 60, 98,
	190, 224, 2, 92, 223, 129, 99, 61, 124, 34, 192, 158, 29, 67, 161, 255,
	70, 24, 250, 164, 39, 121, 155, 197, 132, 218, 56, 102, 229, 187, 89, 7,
	219, 133, 103, 57, 186, 228, 6, 88, 25, 71, 165, 251, 120, 38, 196, 154,
	101, 59, 217, 135, 4, 90, 184, 230, 167, 249, 27, 69, 198, 152, 122, 36,
	248, 166, 68, 26, 153, 199, 37, 123, 58, 100, 134, 216, 91, 5, 231, 185,
	140, 210, 48, 110, 237, 179, 81, 15, 78, 16, 242, 172, 47, 113, 147, 205,
	17, 79, 173, 243, 112, 46, 204, 146, 211, 141, 111, 49, 178, 236, 14, 80,
	175, 241, 19, 77, 206, 144, 114, 44, 109, 51, 209, 143, 12, 82, 176, 238,
	50, 108, 142, 208, 83, 13, 239, 177, 240, 174, 76, 18, 145, 207, 45, 115,
	202, 148, 118, 40, 171, 245, 23, 73, 8, 86, 180, 234, 105, 55, 213, 139,
	87, 9, 235, 181, 54, 104, 138, 212, 149, 203, 41, 119, 244, 170, 72, 22,
	233, 183, 85, 11, 136, 214, 52, 106, 43, 117, 151, 201, 74, 20, 246, 168,
	116, 42, 200, 150, 21, 75, 169, 247, 182, 232, 10, 84, 215, 137, 107, 53,
}
// DS18B20 represents a DS18B20 temperature sensor.
type DS18B20 struct {
	Device embd.W1Device

	Raw    int16
	mu     sync.Mutex
}

// New returns a handle to a DS18B20 sensor.
func New(device embd.W1Device) *DS18B20 {
	return &DS18B20{Device: device}
}

func ds18b20_crc(data []byte) byte {
	var crc byte = 0
	var x byte
	for _, x := range data {
		crc = ds18b20_crc_data[crc ^ x]
	}
	return x
}

func (sensor *DS18B20) ReadTemperature() error {
	sensor.mu.Lock()
	defer sensor.mu.Unlock()

	err := sensor.Device.WriteByte(0x44)

	if err != nil {
		return err
	}

	var ret byte

	for ret, err = sensor.Device.ReadByte(); ret == 0 && err == nil; ret, err = sensor.Device.ReadByte() {}

	if err != nil {
		return err
	}

	err = sensor.Device.WriteByte(0xBE)

	if err != nil {
		return err
	}

	res, err := sensor.Device.ReadBytes(9)

	if err != nil {
		return err
	}

	crc := ds18b20_crc(res[:8])

	if crc != res[8] {
		return errors.New("CRC error")
	}

	sensor.Raw = int16(res[1]) * 256 + int16(res[0])
	cfg := res[4] & 0x60

	switch cfg {
	case 0x00:
		sensor.Raw &^= 7
		break
	case 0x20:
		sensor.Raw &^= 3
		break
	case 0x40:
		sensor.Raw &^= 1
		break
	}

	return nil
}

func (sensor *DS18B20) Celsius() float32 {
	return float32(sensor.Raw) * 0.0625
}

func (sensor *DS18B20) Fahrenheit() float32 {
	return float32(sensor.Raw) * 0.1125 + 32.
}

func (sensor *DS18B20) SetResolution(resolution DS18B20_Resolution) error {
	sensor.mu.Lock()
	defer sensor.mu.Unlock()

	err := sensor.Device.WriteByte(0x4E)
	if err != nil {
		return err
	}
	err = sensor.Device.WriteByte(0x00)
	if err != nil {
		return err
	}
	err = sensor.Device.WriteByte(0x00)
	if err != nil {
		return err
	}
	switch resolution {
	case Resolution_9bit:
		err = sensor.Device.WriteByte(0x1F)
		break
	case Resolution_10bit:
		err = sensor.Device.WriteByte(0x3F)
		break
	case Resolution_11bit:
		err = sensor.Device.WriteByte(0x5F)
		break
	case Resolution_12bit:
		err = sensor.Device.WriteByte(0x7F)
		break
	}
	if err != nil {
		return err
	}
	err = sensor.Device.WriteByte(0x48)
	if err != nil {
		return err
	}
	return nil
}

