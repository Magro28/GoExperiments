package main

//SDL initialisation code with colors ;)

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight int = 800, 600

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

type fileobject struct {
	path string
	size int64
}

func fileWalk(path string) []fileobject {
	var files []fileobject

	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			files = append(files, fileobject{path, info.Size()})
			//fmt.Println(path, info.Size())
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	return files
}

func drawFiles(files []fileobject, pixels []byte, w, h, scale int, blockscale float64) {
	clear(pixels)
	//offset := 5
	var maxSize int64

	for i := range files {
		if files[i].size > maxSize {
			maxSize = files[i].size
		}
	}
	for i := range files {
		filesize := files[i].size
		//fermat spiral x=r(i)*cos i , y=r(i)*sin i }
		x := -1*int(math.Cos(float64(i))*float64(i/scale)) + winWidth/2
		y := int(math.Sin(float64(i))*float64(i/scale)) + winHeight/2
		//calculate radius size
		fileRadius := (float64(255) / float64(maxSize) * float64(filesize) * blockscale)

		//draw circle
		for j := int(-fileRadius); j <= int(fileRadius); j++ {
			for k := int(-fileRadius); k < int(fileRadius); k++ {
				if (k*k + j*j) <= int(fileRadius*fileRadius) {
					setPixel(j+x, k+y, color{byte(filesize % 255), 0, byte(filesize % 250)}, pixels)
				}
			}
		}
	}
}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func main() {

	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("File Galaxy", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
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

	for y := 0; y < winHeight; y++ {
		for x := 0; x < winWidth; x++ {

		}
	}
	keyState := sdl.GetKeyboardState()
	scale := 500
	blockscale := float64(0.06)
	fmt.Println("---------------------")
	fmt.Println("Enter directory path to scan: ")
	filepath := ""
	for filepath == "" {
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		filepath = input.Text()
	}
	fmt.Println("Start scanning: ", filepath)
	files := fileWalk(filepath)
	drawFiles(files, pixels, winWidth, winHeight, scale, blockscale)
	fmt.Println("Use S KEY to zoom in or SHIFT+S to zoom out.")
	fmt.Println("Use B KEY to change block size in or SHIFT+S to descrease.")
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

		}
		if keyState[sdl.SCANCODE_S] != 0 {
			scale = scale + 10*mult
			if scale <= 0 {
				scale = 1
			}
			fmt.Println("Drawing with scale: ", scale)
			drawFiles(files, pixels, winWidth, winHeight, scale, blockscale)
		}
		if keyState[sdl.SCANCODE_B] != 0 {
			blockscale = blockscale + float64(0.01)*float64(mult)

			fmt.Println("Drawing with blockscale: ", blockscale)
			drawFiles(files, pixels, winWidth, winHeight, scale, blockscale)
		}

		tex.Update(nil, pixels, winWidth*4)
		renderer.Copy(tex, nil, nil)
		renderer.Present()
		sdl.Delay(32)
	}
}
