package main

//SDL initialisation code with colors ;)

import (
	"fmt"
	"image/png"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight int = 800, 600

type texture struct {
	pos         pos
	pixels      []byte
	w, h, pitch int
}

type rgba struct {
	r, g, b byte
}

type pos struct {
	x, y float32
}

func setPixel(x, y int, c rgba, pixels []byte) {
	index := (y*winWidth + x) * 4

	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}

}

func (tex *texture) draw(p pos, pixels []byte) {
	for y := 0; y < tex.h; y++ {
		for x := 0; x < tex.w; x++ {
			screenY := y + int(p.y)
			screenX := x + int(p.x)
			if screenX >= 0 && screenX < winWidth && screenY >= 0 && screenY < winHeight {
				texIndex := y*tex.pitch + x*4
				screenIndex := screenY*winWidth*4 + screenX*4

				pixels[screenIndex] = tex.pixels[texIndex]
				pixels[screenIndex+1] = tex.pixels[texIndex+1]
				pixels[screenIndex+2] = tex.pixels[texIndex+2]
				pixels[screenIndex+3] = tex.pixels[texIndex+3]
			}

		}
	}
}

// see https://en.wikipedia.org/wiki/Alpha_compositing
func (tex *texture) drawWithAlphaBlending(pixels []byte) {
	for y := 0; y < tex.h; y++ {
		for x := 0; x < tex.w; x++ {
			screenY := y + int(tex.pos.y)
			screenX := x + int(tex.pos.x)
			if screenX >= 0 && screenX < winWidth && screenY >= 0 && screenY < winHeight {
				texIndex := y*tex.pitch + x*4
				screenIndex := screenY*winWidth*4 + screenX*4

				srcRed := int(tex.pixels[texIndex])
				srcGreen := int(tex.pixels[texIndex+1])
				srcBlue := int(tex.pixels[texIndex+2])
				srcAlpha := int(tex.pixels[texIndex+3])

				dstRed := int(pixels[screenIndex])
				dstGreen := int(pixels[screenIndex+1])
				dstBlue := int(pixels[screenIndex+2])

				resultRed := (srcRed*255 + dstRed*(255-srcAlpha)) / 255
				resultGreen := (srcGreen*255 + dstGreen*(255-srcAlpha)) / 255
				resultBlue := (srcBlue*255 + dstBlue*(255-srcAlpha)) / 255

				pixels[screenIndex] = byte(resultRed)
				pixels[screenIndex+1] = byte(resultGreen)
				pixels[screenIndex+2] = byte(resultBlue)
			}

		}
	}
}

func loadImage(imgpath string) *texture {

	infile, err := os.Open(imgpath)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	img, err := png.Decode(infile)
	if err != nil {
		panic(err)
	}

	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y
	imgPixels := make([]byte, w*h*4)
	imgIndex := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			imgPixels[imgIndex] = byte(r / 256)
			imgIndex++
			imgPixels[imgIndex] = byte(g / 256)
			imgIndex++
			imgPixels[imgIndex] = byte(b / 256)
			imgIndex++
			imgPixels[imgIndex] = byte(a / 256)
			imgIndex++
		}
	}
	return &texture{pos{0, 0}, imgPixels, w, h, w * 4}
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

	window, err := sdl.CreateWindow("2D graphics loading test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
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
	imgTex1 := loadImage("Knight1.png")
	imgTex1.pos = pos{0, 0}
	imgTex1.drawWithAlphaBlending(pixels)

	imgTex2 := loadImage("Knight2.png")
	imgTex2.pos = pos{150, 100}
	imgTex2.drawWithAlphaBlending(pixels)

	imgTex3 := loadImage("Knight3.png")
	imgTex3.pos = pos{500, 200}
	imgTex3.drawWithAlphaBlending(pixels)

	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		tex.Update(nil, pixels, winWidth*4)
		renderer.Copy(tex, nil, nil)
		renderer.Present()

		sdl.Delay(16)
	}

}
