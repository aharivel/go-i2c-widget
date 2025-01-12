# I2C User-Space Drivers for Raspberry Pi Pico

This project leverages the [rp2040-i2c-interface](https://github.com/Nicolai-Electronics/rp2040-i2c-interface) to facilitate communication with various I2C displays and sensors. By combining the Raspberry Pi Pico, the i2c-tiny-usb interface, and Linux, this program provides user-space drivers to interact with connected I2C devices.

> **Note:** This project is experimental, under active development, and intended primarily for learning purposes. Expect frequent changes and potential instability.

---

## Features

- **Integrates with i2c-tiny-usb**: Uses the I2C interface exposed by the Raspberry Pi Pico, programmed with the firmware from the [rp2040-i2c-interface](https://github.com/Nicolai-Electronics/rp2040-i2c-interface) project.
- **User-Space Drivers**: Provides an easy-to-use set of Go-based user-space drivers for interacting with connected I2C devices.
- **Supported I2C Devices**:  
  - **OLED 128x64 Displays** (SSD1306)  
  - **OLED 128x128 Displays** (SH1107)  
  - **Temperature & Humidity Sensor** (SHT31)

---

## How It Works

1. **Flash the Pico**: Program the Raspberry Pi Pico with the [rp2040-i2c-interface](https://github.com/Nicolai-Electronics/rp2040-i2c-interface) firmware to enable it as an I2C device.
2. **Connect I2C Devices**: Attach supported I2C devices (e.g., displays, sensors) to the Pico.
3. **Run This Program**: Use the user-space drivers provided by this project to interact with the connected I2C devices from Linux.

---

## Getting Started

### Prerequisites
- **Hardware**:  
  - Raspberry Pi Pico  
  - I2C devices (e.g., SSD1306, SH1107, SHT31)
- **Software**:  
  - Linux system with the i2c-tiny-usb kernel driver installed  
  - Golang installed (minimum version 1.16 recommended)

### Installation

1. **Clone the Repository**:
   ```bash
   git clone <your-repo-url>
   cd <your-repo-directory>
   ```
2. **Build the Program**:
   ```bash
   go build -o main
   ```

3. **Run the Program**:
   ```bash
   sudo ./main
   ```

---

## Roadmap

- Add support for additional I2C devices.
- Improve stability and reliability.
- Enhance documentation with usage examples.
- Provide precompiled binaries for easier setup.

---

## Contributions

Contributions are welcome! Whether it’s adding support for new devices, fixing bugs, or improving documentation, feel free to open an issue or submit a pull request.

---

## License

This project is licensed under the [ GNU General Public License v3.0.](LICENSE).

---

## Acknowledgments

- [Nicolai-Electronics/rp2040-i2c-interface](https://github.com/Nicolai-Electronics/rp2040-i2c-interface) for the core firmware enabling I2C communication with the Pico.

---

## Disclaimer

This project is highly experimental and may not be suitable for production use. It was created as a learning exercise in Golang and to explore I2C communication.

