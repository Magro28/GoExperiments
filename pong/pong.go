package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight int = 800, 600

type color struct {
	r, g, b byte
}

type pos struct {
	x, y float32
}

type ball struct {
	pos
	radius float32
	xv     float32
	yv     float32
	color
}

func (ball *ball) draw(pixels []byte) {
	for y := -ball.radius; y < ball.radius; y++ {
		for x := -ball.radius; x < ball.radius; x++ {
			if x*x+y*y < ball.radius*ball.radius {
				setPixel(int(ball.x+x), int(ball.y+y), ball.color, pixels)
			}
		}
	}
}

func (ball *ball) update(leftPaddle *paddle, rightPaddle *paddle) {
	ball.x += ball.xv
	ball.y += ball.yv

	//bouncing of the top and bottom
	if ball.y-ball.radius < 0 || ball.y+ball.radius > float32(winHeight) {
		ball.yv = -ball.yv
	}

	if ball.x-ball.radius < 0 || ball.x+ball.radius > float32(winWidth) {
		ball.x = 300
		ball.y = 300
	}

	//left paddle collision
	if ball.x < leftPaddle.x+leftPaddle.w/2 {
		if ball.y > leftPaddle.y-leftPaddle.h/2 && ball.y < leftPaddle.y+leftPaddle.h/2 {
			ball.xv = -ball.xv
		}
	}
	//right paddle collision
	if ball.x > rightPaddle.x-rightPaddle.w/2 {
		if ball.y > rightPaddle.y-rightPaddle.h/2 && ball.y < rightPaddle.y+rightPaddle.h/2 {
			ball.xv = -ball.xv
		}
	}
}

type paddle struct {
	pos
	w float32
	h float32
	color
}

func (paddle *paddle) draw(pixels []byte) {
	startX := paddle.x - paddle.w/2
	startY := paddle.y - paddle.h/2

	for y := 0; y < int(paddle.h); y++ {
		for x := 0; x < int(paddle.w); x++ {
			setPixel(int(startX+float32(x)), int(startY+float32(y)), paddle.color, pixels)
		}
	}
}

func (paddle *paddle) update(keyState []uint8) {
	if keyState[sdl.SCANCODE_UP] != 0 {
		paddle.y--
	}
	if keyState[sdl.SCANCODE_DOWN] != 0 {
		paddle.y++
	}
}

func (paddle *paddle) aiUpdate(ball *ball) {
	paddle.y = ball.y
}

func setPixel(x, y int, c color, pixels []byte) {
	index := (y*winWidth + x) * 4

	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}

}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func main() {

	// Added after EP06 to address macosx issues
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Testing SDL2", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
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

	player1 := paddle{pos{50, 100}, 20, 100, color{255, 255, 255}}
	player2 := paddle{pos{float32(winWidth - 50), 100}, 20, 100, color{255, 255, 255}}

	ball := ball{pos{300, 300}, 20, 2, 2, color{255, 255, 255}}

	//Keyboard Inputs Array
	keyState := sdl.GetKeyboardState()

	// OSX requires that you consume events for windows to open and work properly
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}
		clear(pixels)

		player1.update(keyState)
		player2.aiUpdate(&ball)
		ball.update(&player1, &player2)
		ball.draw(pixels)
		player1.draw(pixels)
		player2.draw(pixels)

		tex.Update(nil, pixels, winWidth*4)
		renderer.Copy(tex, nil, nil)
		renderer.Present()
		sdl.Delay(16)

	}
}