package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	////"golang.org/x/image/font/basicfont"
	"dev/pkg/i2c"
	"dev/pkg/sht31"
	"dev/pkg/ssh1107"

	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

const shellyURL string = "http://192.168.1.126/rpc"

var (
	displayBuffer = make([]byte, 2048)
	mu            sync.Mutex
)

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

type Response struct {
	ID     int    `json:"id"`
	Src    string `json:"src"`
	Result struct {
		ID int     `json:"id"`
		TC float64 `json:"tC"`
		TF float64 `json:"tF"`
	} `json:"result"`
}

func GetTemperatureC(url string) (float64, error) {
	// Prepare the POST request payload
	payload := []byte(`{"id":1,"method":"Temperature.GetStatus","params":{"id":100}}`)

	// Make the POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the JSON response
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Return the temperature in Celsius
	return response.Result.TC, nil
}

func getCirclePoint(cx, cy, radius int, angle float64) (int, int) {
	// Convert angle to radians
	rad := angle * (math.Pi / 180.0)

	// Calculate x and y using cosine and sine
	x := cx + int(float64(radius)*math.Cos(rad))
	y := cy + int(float64(radius)*math.Sin(rad))

	return x, y
}

// BresenhamVector draws a vector of fixed length starting at (x0, y0) and pointing towards (x1, y1).
func BresenhamVector(img draw.Image, color color.Color, x0, y0, x1, y1, length int) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	var sx, sy int
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}
	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}
	err := dx - dy

	// Step counter to track how many pixels have been drawn
	steps := 0

	for {
		// Draw the current pixel
		img.Set(x0, y0, color)
		steps++

		// Stop if we've reached the desired vector length
		if steps >= length {
			return
		}

		// Bresenham's error calculation and stepping
		e2 := 2 * err
		if e2 > -dy {
			err = err - dy
			x0 = x0 + sx
		}
		if e2 < dx {
			err = err + dx
			y0 = y0 + sy
		}
	}
}

// Bresenham draws a line between (x0, y0) and (x1, y1)
func Bresenham(img draw.Image, color color.Color, x0, y0, x1, y1 int) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	var sx, sy int
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}
	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}
	err := dx - dy

	var e2 int
	for {
		img.Set(x0, y0, color)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 = 2 * err
		if e2 > -dy {
			err = err - dy
			x0 = x0 + sx
		}
		if e2 < dx {
			err = err + dx
			y0 = y0 + sy
		}
	}
}

func pause() {
	fmt.Println("Press any key to continue...")
	var input string
	fmt.Scanf("%s", &input)
}

//==============================================================================

type ClockScreen struct {
	display ssh1107.Display
	mu      *sync.RWMutex
	ang     float64
	angle   float64
}

func (cs *ClockScreen) Draw() {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	cs.display.ClearImage(color.RGBA{0, 0, 0, 0})
	img := cs.display.GetImage()
	for cs.ang = 0.0; cs.ang <= 360.0; cs.ang = cs.ang + 30 {
		cx, cy := getCirclePoint(64, 64, 63, cs.ang)
		BresenhamVector(img, color.White, cx, cy, 64, 64, 10)

	}

	cs.display.DrawCircle(64, 64, 63)
	cx, cy := getCirclePoint(64, 64, 63, cs.angle)

	Bresenham(img, color.White, 64, 64, cx, cy)
	cs.display.Draw()
	cs.display.Display_old()
	cs.display.DisplayOn()

	Bresenham(img, color.Black, 64, 64, cx, cy)

	// cs.display.Clear()
}

func (cs *ClockScreen) Update() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.angle = cs.angle + 6.0
	buffer := cs.display.GetBuffer()
	mu.Lock()
	defer mu.Unlock()
	copy(displayBuffer, buffer)
}

// ==============================================================================
type LogoScreen struct {
	display ssh1107.Display
}

func (ls *LogoScreen) Draw() {
	mu.Lock()
	defer mu.Unlock()

	ls.display.DrawImg()
	ls.display.Display_old()
	ls.display.DisplayOn()
}

func (ls *LogoScreen) Update() {
	buffer := ls.display.GetBuffer()
	mu.Lock()
	defer mu.Unlock()

	copy(displayBuffer, buffer)
}

// ==============================================================================
type ShellyScreen struct {
	display     ssh1107.Display
	temperature float64
	mu          *sync.RWMutex
}

func (ss *ShellyScreen) Draw() {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	img := ss.display.GetImage()

	ss.display.ClearImage(color.RGBA{0, 0, 0, 0})

	// Define points for each line (start and end)
	x0, y0 := 0, 0 // Top-left to top-right
	x1, y1 := img.Bounds().Dx()-1, 0

	Bresenham(img, color.White, x0, y0, x1, y1)

	// Second line: top-right to bottom-right
	x0, y0 = img.Bounds().Dx()-1, 0
	x1, y1 = img.Bounds().Dx()-1, img.Bounds().Dy()-1

	Bresenham(img, color.White, x0, y0, x1, y1)

	// Third line: bottom-right to bottom-left
	x0, y0 = img.Bounds().Dx()-1, img.Bounds().Dy()-1
	x1, y1 = 0, img.Bounds().Dy()-1

	Bresenham(img, color.White, x0, y0, x1, y1)

	// Fourth line: bottom-left to top-left
	x0, y0 = 0, img.Bounds().Dy()-1
	x1, y1 = 0, 0

	Bresenham(img, color.White, x0, y0, x1, y1)

	// Set the starting point for drawing text
	drawer := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{color.White},
		Face: inconsolata.Bold8x16,
		Dot:  fixed.Point26_6{},
	}
	// Set the starting point for drawing text
	drawer.Dot = fixed.Point26_6{
		X: fixed.Int26_6(16 * 64),
		Y: fixed.Int26_6(32 * 64),
	}
	// Draw the title
	drawer.DrawString("Shelly")

	// Set the starting point for drawing text
	drawer = &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{color.White},
		Face: inconsolata.Regular8x16,
		Dot:  fixed.Point26_6{},
	}
	// Set the starting point for drawing text
	drawer.Dot = fixed.Point26_6{
		X: fixed.Int26_6(16 * 64),
		Y: fixed.Int26_6(64 * 64),
	}
	// Draw the temperature string below the title
	temp_string := fmt.Sprintf("Temp: %.2f °C", ss.temperature)
	drawer.DrawString(temp_string)

	ss.display.Draw()
	ss.display.Display_old()
	ss.display.DisplayOn()
}

func (ss *ShellyScreen) Update() {
	var err error = nil

	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.temperature, err = GetTemperatureC(shellyURL)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	buffer := ss.display.GetBuffer()
	mu.Lock()
	defer mu.Unlock()
	copy(displayBuffer, buffer)
}

//==============================================================================

type Screen interface {
	Draw()
	Update()
}

type ScreenManager struct {
	screens      []Screen // List of screens
	currentIndex int      // Active screen index
}

func (sm *ScreenManager) CurrentScreen() Screen {
	return sm.screens[sm.currentIndex]
}

func (sm *ScreenManager) NextScreen() {
	sm.currentIndex = (sm.currentIndex + 1) % len(sm.screens)
}

func (sm *ScreenManager) PreviousScreen() {
	sm.currentIndex = (sm.currentIndex - 1 + len(sm.screens)) % len(sm.screens)
}

func (sm *ScreenManager) StartUpdating() {
	for _, screen := range sm.screens {
		screen.Update()
	}
}

// Timer Goroutine
func startTimers(updateChan, switchScreenChan chan bool) {
	updateTicker := time.NewTicker(1 * time.Second)
	switchScreenTicker := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-updateTicker.C:
			updateChan <- true
		case <-switchScreenTicker.C:
			switchScreenChan <- true
		}
	}
}

// ==============================================================================
func serveBuffer(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	// Convert []byte to a slice of interface{}
	var array []interface{}
	for _, b := range displayBuffer {
		array = append(array, b)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(array)
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "oled.html")
}

// ==============================================================================
func main() {
	fmt.Println("### init server... ")
	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/buffer", serveBuffer)
	go http.ListenAndServe(":8088", nil)

	// Initialize I2C SHT31 device
	sht31_dev, err := i2c.Init(9, 0x44) // bus 1, address 0x3c
	if err != nil {
		log.Fatalf("Failed to initialize I2C device: %v", err)
	}
	defer sht31_dev.Close()
	// Initialize I2C sh1107 device
	ssh1107_dev, err := i2c.Init(9, 0x3c) // bus 1, address 0x3c
	if err != nil {
		log.Fatalf("Failed to initialize I2C device: %v", err)
	}
	defer ssh1107_dev.Close()

	sensor := sht31.NewSHT31(sht31_dev)

	temp := sensor.ReadTemperature()
	hum := sensor.ReadHumidity()

	fmt.Printf("Temperature: %.2f °C, Humidity: %.2f %%\n", temp, hum)

	fmt.Println("...NewDisplay...")
	// Create a Display from a I2c Device & a Screen
	display, err := ssh1107.NewDisplay(ssh1107_dev, ssh1107.NewScreen(128, 128))
	if err != nil {
		log.Fatalf("SH1107 display init failed: %v", err)
	}
	// Initialize a timer for automatic screen switching
	autoSwitchTicker := time.NewTicker(10 * time.Second)
	defer autoSwitchTicker.Stop()
	// Rendering refresh rate (e.g., 30 FPS = ~33ms per frame)
	renderTicker := time.NewTicker(33 * time.Millisecond)
	defer renderTicker.Stop()

	// Create screens
	logoScreen := &LogoScreen{display: display}
	clockScreen := &ClockScreen{display: display, mu: &sync.RWMutex{}}
	shellyScreen := &ShellyScreen{display: display, mu: &sync.RWMutex{}}

	// Create screen manager
	screenManager := &ScreenManager{
		screens:      []Screen{logoScreen, clockScreen, shellyScreen},
		currentIndex: 0,
	}

	// Channels for events
	updateChan := make(chan bool)
	switchScreenChan := make(chan bool)

	// Start timer goroutine
	go startTimers(updateChan, switchScreenChan)
	fmt.Println("...Initialize()...")
	display.Initialize()

	// Main loop
	for {
		select {
		case <-updateChan:
			// Update all screens
			screenManager.StartUpdating()
		case <-switchScreenChan:
			// Switch to the next screen
			screenManager.NextScreen()
		default:
			// Render the current screen
			screen := screenManager.CurrentScreen()
			screen.Draw()                     // render the current screen
			time.Sleep(16 * time.Millisecond) // Roughly 60 FPS
		}
	}
}
