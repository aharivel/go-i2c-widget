package ssd1306

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"dev/pkg/i2c"
)

const (
	OLED_CMD                 = 0x80
	OLED_CMD_COL_ADDRESSING  = 0x21
	OLED_CMD_PAGE_ADDRESSING = 0x22
	OLED_CMD_CONTRAST        = 0x81
	OLED_CMD_START_COLUMN    = 0x00
	OLED_CMD_HIGH_COLUMN     = 0x10
	OLED_CMD_DISPLAY_OFF     = 0xAE
	OLED_CMD_DISPLAY_ON      = 0xAF
	OLED_DATA                = 0x40
	OLED_ADRESSING           = 0x21
	OLED_ADRESSING_START     = 0xB0
	OLED_ADRESSING_COL       = 0x21
	OLED_END                 = 0x10
	PIXSIZE                  = 8

	SSD1306_CMD                 = 0x80
	SSD1306_SETDISPLAYCLOCKDIV  = 0xD5
	SSD1306_DISPLAYOFF          = 0xAE
	SSD1306_SETMULTIPLEX        = 0xA8
	SSD1306_SETDISPLAYOFFSET    = 0xD3
	SSD1306_SETSTARTLINE        = 0x0
	SSD1306_CHARGEPUMP          = 0x8D
	SSD1306_MEMORYMODE          = 0x20
	SSD1306_SEGREMAP            = 0xA0
	SSD1306_COMSCANDEC          = 0xC8
	SSD1306_SETCOMPINS          = 0xDA
	SSD1306_SETCONTRAST         = 0x81
	SSD1306_SETPRECHARGE        = 0xD9
	SSD1306_SETVCOMDETECT       = 0xDB
	SSD1306_DISPLAYALLON_RESUME = 0xA4
	SSD1306_NORMALDISPLAY       = 0xA6
	SSD1306_EXTERNALVCC         = true
	SSD1306_SWITCHCAPVCC        = false
)

type Display interface {
	Initialize() error
	DisplayOn() (int, error)
	DisplayOff() (int, error)
	Display(*screen) error
	Clear(*screen)
	Draw(*screen)
	DrawPix(*screen, int, int)
	ClearImage(*screen, color.Color)
}

type SSD1306_128_64 struct {
	fd       *i2c.I2CDevice
	vccstate bool
}

// Struct for representing screen properties
type screen struct {
	h        int
	w        int
	contrast int
	buffer   []byte
	Img      draw.Image
}

// ///////////////////////////////////////////////////////
//
// # Declaration Functions
//
// //////////////////////////////////////////////////////
func NewScreen(h, w int, vccState bool) *screen {
	return &screen{
		h:   h,
		w:   w,
		Img: image.NewRGBA((image.Rect(0, 0, int(w), int(h)))),
	}
}

func NewDisplay(fd *i2c.I2CDevice, ecran *screen) (Display, error) {
	fmt.Println("Init display", ecran.w, ecran.h)
	switch {
	case ecran.w == 128 && ecran.h == 32:
		return nil, fmt.Errorf("unsupported display h=32")
	case ecran.w == 128 && ecran.h == 64:
		return newSSD1306_128_64(fd, false), nil
	case ecran.w == 96 && ecran.h == 16:
		return nil, fmt.Errorf("unsupported display h=16")
	default:
		return nil, fmt.Errorf("unsupported display dimensions: %dx%d", ecran.w, ecran.h)
	}
}

/////////////////////////////////////////////////////////
//
// Interface Functions
//
////////////////////////////////////////////////////////

func (d *SSD1306_128_64) Initialize() error {
	fmt.Println("Initialize screen")

	data := []byte{
		SSD1306_DISPLAYOFF,         // 0xAE
		SSD1306_SETDISPLAYCLOCKDIV, // 0xD5
		0x80,                       // the suggested ratio 0x80
		SSD1306_SETMULTIPLEX,       // 0xA8
		0x3F,                       // Multiplex value for 128x64
		SSD1306_SETDISPLAYOFFSET,   // 0xD3
		0x0,                        // no offset
		SSD1306_SETSTARTLINE | 0x0, // line #0
		SSD1306_CHARGEPUMP,         // 0x8D
	}

	// Adjust charge pump settings based on vccstate.
	if d.vccstate == SSD1306_EXTERNALVCC {
		data = append(data, byte(0x10)) // External Vcc
	} else {
		data = append(data, byte(0x14)) // Internal Vcc
	}

	// Additional setup commands.
	data = append(data, []byte{
		SSD1306_MEMORYMODE,     // 0x20
		0x00,                   // 0x0 act like ks0108
		SSD1306_SEGREMAP | 0x1, // Map segment 0 to column 127
		SSD1306_COMSCANDEC,     // Scan in descending order
		SSD1306_SETCOMPINS,     // 0xDA
		0x12,                   // Sequential COM pin configuration for 128x64
		SSD1306_SETCONTRAST,    // 0x81
	}...)

	// Set contrast based on vccstate.
	if d.vccstate == SSD1306_EXTERNALVCC {
		data = append(data, byte(0x9F)) // Contrast value for External Vcc
	} else {
		data = append(data, byte(0xCF)) // Contrast value for Internal Vcc
	}

	// More commands, including setting the precharge period.
	data = append(data, SSD1306_SETPRECHARGE) // 0xd9

	if d.vccstate == SSD1306_EXTERNALVCC {
		data = append(data, byte(0x22)) // Precharge value for External Vcc
	} else {
		data = append(data, byte(0xF1)) // Precharge value for Internal Vcc
	}

	// Final setup commands.
	data = append(data, []byte{
		SSD1306_SETVCOMDETECT,       // 0xDB
		0x40,                        // VCOM deselect level
		SSD1306_DISPLAYALLON_RESUME, // 0xA4
		SSD1306_NORMALDISPLAY,       // 0xA6
	}...)

	return sendCommands(*d.fd, data...)
}

// Turn on OLED display
func (d *SSD1306_128_64) DisplayOn() (int, error) {
	return writeCommand(*d.fd, OLED_CMD_DISPLAY_ON)
}

// Turn off OLED display
func (d *SSD1306_128_64) DisplayOff() (int, error) {
	return writeCommand(*d.fd, OLED_CMD_DISPLAY_OFF)
}

// Display buffer to the screen
func (d *SSD1306_128_64) Display(ecran *screen) error {
	writeCommand(*d.fd, OLED_CMD_COL_ADDRESSING) //
	writeCommand(*d.fd, 0)
	writeCommand(*d.fd, byte(ecran.w-1))
	writeCommand(*d.fd, OLED_CMD_PAGE_ADDRESSING) //
	writeCommand(*d.fd, 0)
	writeCommand(*d.fd, byte((ecran.h/8)-1))

	for i := 0; i < len(ecran.buffer); i += 64 {
		data := ecran.buffer[i : i+64]
		_, err := writeData(*d.fd, data)
		// fmt.Println(data) //check RAM
		if err != nil {
			return err
		}
	}

	return nil
}

// Clear the OLED screen
func (d *SSD1306_128_64) Clear(ecran *screen) {
	size := ecran.w * ecran.h / PIXSIZE
	ecran.buffer = make([]byte, size)
}

func (d *SSD1306_128_64) ClearImage(ecran *screen, col color.Color) {
	// Fill the entire image with the specified color
	draw.Draw(ecran.Img, ecran.Img.Bounds(), &image.Uniform{C: col}, image.Point{}, draw.Src)
}

// Set pixel
func (d *SSD1306_128_64) DrawPix(ecran *screen, x int, y int) {
	if x >= 0 && x < ecran.Img.Bounds().Dx() && y >= 0 && y < ecran.Img.Bounds().Dy() {
		ecran.Img.Set(x, y, color.White)
	}
}

func (d *SSD1306_128_64) convertImageToOLEDData(ecran *screen) ([]byte, error) {
	bounds := ecran.Img.Bounds()
	if bounds.Max.X != ecran.w || ecran.h != bounds.Max.Y {
		panic(fmt.Sprintf("Error: Size of image is not %dx%d pixels.", ecran.w, ecran.h))
	}
	size := ecran.w * ecran.h / PIXSIZE
	data := make([]byte, size)
	for page := 0; page < ecran.h/8; page++ {
		for x := 0; x < ecran.w; x++ {
			bits := uint8(0)
			for bit := 0; bit < 8; bit++ {
				y := page*8 + 7 - bit
				if y < ecran.h {
					col := color.GrayModel.Convert(ecran.Img.At(x, y)).(color.Gray)
					if col.Y > 127 {
						bits = (bits << 1) | 1
					} else {
						bits = bits << 1
					}
				}
			}
			index := page*ecran.w + x
			data[index] = byte(bits)
		}
	}
	return data, nil
}

func (d *SSD1306_128_64) Draw(ecran *screen) {
	ecran.buffer, _ = d.convertImageToOLEDData(ecran)
}

// ///////////////////////////////////////////////////////
//
// # Private Functions
//
// //////////////////////////////////////////////////////
func newSSD1306_128_64(fd *i2c.I2CDevice, vccstate bool) *SSD1306_128_64 {
	return &SSD1306_128_64{
		fd:       fd,
		vccstate: vccstate,
	}
}

// Send data to OLED
func writeData(fd i2c.I2CDevice, data []byte) (int, error) {
	res := 0
	for _, value := range data {
		if _, err := fd.Write([]byte{OLED_DATA, value}); err != nil {
			return res, err
		}
		res++
	}
	return res, nil
}

// writeCommand sends a single command byte to the SSD1306 device.
func writeCommand(fd i2c.I2CDevice, cmd byte) (int, error) {
	return fd.Write([]byte{SSD1306_CMD, cmd})
}

// sendCommands sends a sequence of command bytes to the SSD1306 device.
func sendCommands(fd i2c.I2CDevice, commands ...byte) error {
	for _, cmd := range commands {
		if _, err := writeCommand(fd, cmd); err != nil {
			return err
		}
	}
	return nil
}
