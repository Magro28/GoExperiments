package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight int = 800, 600

func lerp(b1 byte, b2 byte, pct float32) byte {
	return uint8(float32(b1) + pct*(float32(b2)-float32(b1)))
}

func colorLerp(c1, c2 color, pct float32) color {
	return color{lerp(c1.r, c2.r, pct), lerp(c1.g, c2.g, pct), lerp(c1.b, c2.b, pct)}
}

func getGradient(c1, c2 color) []color {
	result := make([]color, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		result[i] = colorLerp(c1, c2, pct)
	}
	return result
}

//get multi color gradient (first c1, c2 then switch to c3, c4)
func getDualGradient(c1, c2, c3, c4 color) []color {
	result := make([]color, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		if pct < 0.5 {
			result[i] = colorLerp(c1, c2, pct*float32(2))
		} else {
			result[i] = colorLerp(c3, c4, pct*float32(1.5)-float32(0.5))
		}
	}
	return result
}

func clamp(min, max, v int) int {
	if v < min {
		v = min
	} else if v > max {
		v = max
	}
	return v
}

func rescaleAndDraw(noise []float32, min, max float32, gradient []color, pixels []byte) {
	scale := 255 / (max - min)
	offset := min * scale

	for i := range noise {
		noise[i] = noise[i]*scale - offset
		c := gradient[clamp(0, 255, int(noise[i]))]
		p := i * 4
		pixels[p] = c.r
		pixels[p+1] = c.g
		pixels[p+2] = c.b
	}
}

//fractal noise
func fbm2(x, y, frequency, lacunarity, gain float32, octaves int) float32 {
	var sum float32
	amplitude := float32(1.0)
	for i := 0; i < octaves; i++ {
		sum += snoise2(x*frequency, y*frequency) * amplitude
		frequency = frequency * lacunarity
		amplitude = amplitude * gain
	}
	return sum
}

func makeNoise(pixels []byte, frequency, lacunarity, gain float32, octaves int) {

	noise := make([]float32, winWidth*winHeight)
	fmt.Println("frequency", frequency, "lacunarity", lacunarity, "gain", gain, "octaves", octaves)
	i := 0
	min := float32(9999.0)
	max := float32(-9999.9)
	for y := 0; y < winHeight; y++ {
		for x := 0; x < winWidth; x++ {
			noise[i] = fbm2(float32(x), float32(y), frequency, lacunarity, gain, octaves)

			if noise[i] < min { //water
				min = noise[i]

			} else if noise[i] > max { //islands
				max = noise[i]
			}
			i++
		}
	}
	gradient := getDualGradient(color{0, 0, 175}, color{80, 160, 244}, color{12, 192, 75}, color{255, 255, 255})
	rescaleAndDraw(noise, min, max, gradient, pixels)
}

type color struct {
	r, g, b byte
}

func setPixel(x, y int, c color, pixels []byte) {
	index := (y*winWidth + x) * 4

	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}

}

func main() {

	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Procedural Generation", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer renderer.Destroy()

	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(winWidth), int32(winHeight))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tex.Destroy()

	pixels := make([]byte, winWidth*winHeight*4)

	frequency := float32(0.01)
	gain := float32(0.2)
	lacunarity := float32(3.0)
	octaves := 3
	keyState := sdl.GetKeyboardState()
	makeNoise(pixels, frequency, lacunarity, gain, octaves)
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		mult := 1
		if keyState[sdl.SCANCODE_LSHIFT] != 0 || keyState[sdl.SCANCODE_RSHIFT] != 0 {
			mult = -1
			makeNoise(pixels, frequency, lacunarity, gain, octaves)
		}
		if keyState[sdl.SCANCODE_O] != 0 {
			octaves = octaves + 1*mult
			makeNoise(pixels, frequency, lacunarity, gain, octaves)
		}
		if keyState[sdl.SCANCODE_F] != 0 {
			frequency = frequency + 0.001*float32(mult)
			makeNoise(pixels, frequency, lacunarity, gain, octaves)
		}
		if keyState[sdl.SCANCODE_G] != 0 {
			gain = gain + 0.1*float32(mult)
			makeNoise(pixels, frequency, lacunarity, gain, octaves)
		}
		if keyState[sdl.SCANCODE_L] != 0 {
			lacunarity = lacunarity + 0.1*float32(mult)
			makeNoise(pixels, frequency, lacunarity, gain, octaves)
		}

		tex.Update(nil, pixels, winWidth*4)
		renderer.Copy(tex, nil, nil)
		renderer.Present()
		sdl.Delay(32)
	}

}
