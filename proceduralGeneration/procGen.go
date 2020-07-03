package main

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

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

func getTripleGradient(c1, c2, c3, c4, c5, c6 color) []color {
	result := make([]color, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		if pct < 0.65 {
			result[i] = colorLerp(c1, c2, pct*float32(2))
		} else if pct >= 0.65 && pct < 0.70 {
			result[i] = colorLerp(c3, c4, pct*float32(1.5)-float32(0.5))
		} else {
			result[i] = colorLerp(c5, c6, pct*float32(1.5)-float32(0.5))
		}
	}
	return result
}

func getQuadrupleGradient(c1, c2, c3, c4, c5, c6, c7, c8 color) []color {
	result := make([]color, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		if pct < 0.65 {
			result[i] = colorLerp(c1, c2, pct*float32(2))
		} else if pct >= 0.65 && pct < 0.70 {
			result[i] = colorLerp(c3, c4, pct*float32(1.5)-float32(0.5))
		} else if pct >= 0.7 && pct < 0.90 {
			result[i] = colorLerp(c5, c6, pct*float32(1.5)-float32(0.5))
		} else {
			result[i] = colorLerp(c7, c8, pct*float32(3.5)-float32(0))
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

//turbulence noise
func turbulence(x, y, frequency, lacunarity, gain float32, octaves int) float32 {
	var sum float32
	amplitude := float32(1)

	for i := 0; i < octaves; i++ {
		f := snoise2(x*frequency, y*frequency) * amplitude
		if f < 0 {
			f = -1.0 * f
		}
		sum += f
		frequency = frequency * lacunarity
		amplitude = amplitude * gain
	}
	return sum
}

func makeNoise(pixels []byte, frequency, lacunarity, gain float32, octaves int, colormode int, algorithm int, w, h int) {
	startTime := time.Now()
	noise := make([]float32, winWidth*winHeight)
	fmt.Println("frequency", frequency, "lacunarity", lacunarity, "gain", gain, "octaves", octaves, "colormode", colormode, "algorithm", algorithm)
	var mutex = &sync.Mutex{}
	min := float32(math.MaxFloat32)
	max := float32(-math.MaxFloat32)

	//specify routines, waitgroup and batchsize after logical CPU cores
	numRoutines := runtime.NumCPU()
	var wg sync.WaitGroup
	//wait for x routines to finish
	wg.Add(numRoutines)
	batchSize := len(noise) / numRoutines
	for i := 0; i < numRoutines; i++ {
		//start new thread
		go func(i int) {
			//defer executes after surrounding block
			defer wg.Done()

			innerMin := float32(math.MaxFloat32)
			innerMax := float32(-math.MaxFloat32)

			start := i * batchSize
			end := start + batchSize - 1
			for j := start; j < end; j++ {
				x := j % w
				y := (j - x) / h
				if algorithm <= 1 {
					noise[j] = fbm2(float32(x), float32(y), frequency, lacunarity, gain, octaves)
				} else {
					noise[j] = turbulence(float32(x), float32(y), frequency, lacunarity, gain, octaves)
				}

				if noise[j] < innerMin {
					innerMin = noise[j]

				} else if noise[j] > innerMax {
					innerMax = noise[j]
				}

			}

			mutex.Lock()
			if innerMin < min {
				min = innerMin
			}
			if innerMax > max {
				max = innerMax
			}
			mutex.Unlock()

		}(i)
	}
	//wait for all go routines to finish
	wg.Wait()
	elapseTime := time.Since(startTime).Seconds() * 1000.0
	fmt.Println("Elapsed Time:", elapseTime)

	var gradient []color
	if colormode <= 1 {
		gradient = getGradient(color{255, 0, 0}, color{255, 242, 0})
	} else if colormode == 2 {
		gradient = getDualGradient(color{0, 0, 175}, color{80, 100, 200}, color{12, 192, 75}, color{255, 255, 255})
	} else if colormode == 3 {
		gradient = getTripleGradient(color{0, 0, 100}, color{80, 100, 200}, color{210, 210, 0}, color{100, 80, 0}, color{50, 160, 20}, color{0, 80, 10})
	} else {
		gradient = getQuadrupleGradient(color{0, 0, 100}, color{80, 100, 200}, color{210, 210, 0}, color{100, 80, 0}, color{50, 160, 20}, color{0, 80, 10}, color{100, 100, 100}, color{250, 250, 250})
	}
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
	algorithm := 1
	colormode := 4
	keyState := sdl.GetKeyboardState()
	makeNoise(pixels, frequency, lacunarity, gain, octaves, colormode, algorithm, winWidth, winHeight)
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
			makeNoise(pixels, frequency, lacunarity, gain, octaves, colormode, algorithm, winWidth, winHeight)
		}
		if keyState[sdl.SCANCODE_O] != 0 {
			octaves = octaves + 1*mult
			makeNoise(pixels, frequency, lacunarity, gain, octaves, colormode, algorithm, winWidth, winHeight)
		}
		if keyState[sdl.SCANCODE_F] != 0 {
			frequency = frequency + 0.001*float32(mult)
			makeNoise(pixels, frequency, lacunarity, gain, octaves, colormode, algorithm, winWidth, winHeight)
		}
		if keyState[sdl.SCANCODE_G] != 0 {
			gain = gain + 0.1*float32(mult)
			makeNoise(pixels, frequency, lacunarity, gain, octaves, colormode, algorithm, winWidth, winHeight)
		}
		if keyState[sdl.SCANCODE_L] != 0 {
			lacunarity = lacunarity + 0.1*float32(mult)
			makeNoise(pixels, frequency, lacunarity, gain, octaves, colormode, algorithm, winWidth, winHeight)
		}
		if keyState[sdl.SCANCODE_C] != 0 {
			colormode = colormode + 1*mult
			colormode = clamp(1, 4, colormode)
			makeNoise(pixels, frequency, lacunarity, gain, octaves, colormode, algorithm, winWidth, winHeight)
		}
		if keyState[sdl.SCANCODE_A] != 0 {
			algorithm = algorithm + 1*mult
			algorithm = clamp(1, 2, algorithm)
			makeNoise(pixels, frequency, lacunarity, gain, octaves, colormode, algorithm, winWidth, winHeight)
		}

		tex.Update(nil, pixels, winWidth*4)
		renderer.Copy(tex, nil, nil)
		renderer.Present()
		sdl.Delay(32)
	}

}
