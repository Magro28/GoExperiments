package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight int = 800, 600

//enum definition
type gameState int

const (
	start gameState = iota
	play
)

var state = start

//Number font
var nums = [][]byte{
	{
		1, 1, 1,
		1, 0, 1,
		1, 0, 1,
		1, 0, 1,
		1, 1, 1,
	},
	{
		1, 1, 0,
		0, 1, 0,
		0, 1, 0,
		0, 1, 0,
		0, 1, 0,
	},
	{
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
		1, 0, 0,
		1, 1, 1,
	},
	{
		1, 1, 1,
		0, 0, 1,
		1, 1, 0,
		0, 0, 1,
		1, 1, 1,
	},
}

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

func (ball *ball) update(leftPaddle *paddle, rightPaddle *paddle, elapsedTime float32) {
	ball.x += ball.xv * elapsedTime
	ball.y += ball.yv * elapsedTime

	//bouncing of the top and bottom
	if ball.y-ball.radius < 0 {
		ball.yv = -ball.yv
		//set ball outside of top
		ball.y = ball.radius
	} else if ball.y+ball.radius > float32(winHeight) {
		ball.yv = -ball.yv
		//set ball outside of bottom
		ball.y -= ball.radius
	}

	if ball.x < 0 {
		rightPaddle.score++
		ball.pos = getCenter()
		ball.xv = -400
		ball.yv = 0
		state = start
	} else if ball.x > float32(winWidth) {
		leftPaddle.score++
		ball.pos = getCenter()
		ball.xv = 400
		ball.yv = 0
		state = start
	}

	//left paddle collision
	if ball.x-ball.radius < leftPaddle.x+leftPaddle.w/2 {
		if ball.y > leftPaddle.y-leftPaddle.h/2 && ball.y < leftPaddle.y+leftPaddle.h/2 {
			xyModifier := float32(0.0)
			if ball.y > leftPaddle.y-leftPaddle.h/2 {
				xyModifier = (ball.y/leftPaddle.y - 1) * 1000
			} else {
				xyModifier = (ball.y / leftPaddle.y) * 1000
			}
			ball.xv = -ball.xv
			ball.yv = (ball.yv + float32(xyModifier))

			//set ball outside of paddle
			ball.x = leftPaddle.x + leftPaddle.w/2.0 + ball.radius
		}
	}
	//right paddle collision
	if ball.x+ball.radius > rightPaddle.x-rightPaddle.w/2 {
		if ball.y > rightPaddle.y-rightPaddle.h/2 && ball.y < rightPaddle.y+rightPaddle.h/2 {
			xyModifier := float32(0.0)
			if ball.y > rightPaddle.y-rightPaddle.h/2 {
				xyModifier = (ball.y/rightPaddle.y - 1) * 1000
			} else {
				xyModifier = (ball.y / rightPaddle.y) * 1000
			}
			ball.xv = -ball.xv
			ball.yv = (ball.yv + float32(xyModifier))

			//set ball outside of paddle
			ball.x = rightPaddle.x - rightPaddle.w/2.0 - ball.radius
		}

	}

	//prevend laser balls
	if ball.yv > 1000 {
		ball.yv = 1000
	}
}

func drawNumber(pos pos, color color, size int, num int, pixels []byte) {
	startX := int(pos.x) - (size*3)/2
	startY := int(pos.y) - (size*5)/2

	for i, v := range nums[num] {
		if v == 1 {
			for y := startY; y < startY+size; y++ {
				for x := startX; x < startX+size; x++ {
					setPixel(x, y, color, pixels)
				}
			}
		}
		startX += size
		if (i+1)%3 == 0 {
			startY += size
			startX -= size * 3
		}
	}
}

func getCenter() pos {
	return pos{float32(winWidth / 2), float32(winHeight / 2)}
}

//get you the percentage point of the given range
func lerp(a float32, b float32, pct float32) float32 {
	return (a + pct*(b-a))
}

type paddle struct {
	pos
	w float32
	h float32
	color
	speed float32
	score int
}

func (paddle *paddle) draw(pixels []byte) {
	startX := paddle.x - paddle.w/2
	startY := paddle.y - paddle.h/2

	for y := 0; y < int(paddle.h); y++ {
		for x := 0; x < int(paddle.w); x++ {
			setPixel(int(startX+float32(x)), int(startY+float32(y)), paddle.color, pixels)
		}
	}

	numX := lerp(paddle.x, getCenter().x, 0.2)
	drawNumber(pos{numX, 35}, paddle.color, 10, paddle.score, pixels)
}

func (paddle *paddle) update(keyState []uint8, controllerAxis int16, elapsedTime float32) {
	if keyState[sdl.SCANCODE_UP] != 0 {
		paddle.y -= paddle.speed * elapsedTime
	}
	if keyState[sdl.SCANCODE_DOWN] != 0 {
		paddle.y += paddle.speed * elapsedTime
	}

	//analog stick movement
	//dead point < 1500
	if math.Abs(float64(controllerAxis)) > 1500 {
		//calculate percentage of analog stick press
		pct := float32(controllerAxis) / 32767.0
		paddle.y += paddle.speed * pct * elapsedTime
	}
}

func (paddle *paddle) aiUpdate(ball *ball, elapsedTime float32) {
	if paddle.y < ball.y {
		paddle.y += paddle.speed * elapsedTime
	} else {
		paddle.y -= paddle.speed * elapsedTime
	}
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

	window, err := sdl.CreateWindow("Pong", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
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

	var controllerHandlers []*sdl.GameController
	for i := 0; i < sdl.NumJoysticks(); i++ {
		controllerHandlers = append(controllerHandlers, sdl.GameControllerOpen(i))
		defer controllerHandlers[i].Close()
	}

	//initialize randomizer
	rand.Seed(time.Now().UnixNano())

	pixels := make([]byte, winWidth*winHeight*4)

	player1 := paddle{pos{50, 100}, 20, 100, color{0, 0, 255}, 300, 0}
	player2 := paddle{pos{float32(winWidth - 50), 100}, 20, 100, color{255, 0, 0}, 300, 0}

	ball := ball{getCenter(), 20, 400, 400, color{0, 255, 0}}

	//Keyboard Inputs Array
	keyState := sdl.GetKeyboardState()

	//framerate should be same for each cpu
	var frameStart time.Time
	var elapsedTime float32
	var controllerAxis int16
	// OSX requires that you consume events for windows to open and work properly
	for {
		frameStart = time.Now()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		//get controller input
		for _, controller := range controllerHandlers {
			if controller != nil {
				controllerAxis = controller.Joystick().Axis(sdl.CONTROLLER_AXIS_LEFTY)
			}
		}

		if state == play {
			player1.update(keyState, controllerAxis, elapsedTime)
			player2.aiUpdate(&ball, elapsedTime)
			ball.update(&player1, &player2, elapsedTime)
		} else if state == start {

			//win effect
			if player1.score == 3 {
				player1.color = color{byte(rand.Int() * 255), byte(rand.Int() * 255), byte(rand.Int() * 255)}
			} else if player2.score == 3 {
				player2.color = color{byte(rand.Int() * 255), byte(rand.Int() * 255), byte(rand.Int() * 255)}

			}

			//continue game
			if keyState[sdl.SCANCODE_SPACE] != 0 {
				if player1.score == 3 || player2.score == 3 {
					player1.color = color{0, 0, 255}
					player2.color = color{255, 0, 0}
					player1.score = 0
					player2.score = 0
				}
				state = play
			}
		}
		clear(pixels)
		ball.draw(pixels)
		player1.draw(pixels)
		player2.draw(pixels)

		tex.Update(nil, pixels, winWidth*4)
		renderer.Copy(tex, nil, nil)
		renderer.Present()
		elapsedTime = float32(time.Since(frameStart).Seconds())
		//max 200 fps
		if elapsedTime < 0.005 {
			sdl.Delay(5 - uint32(elapsedTime/1000.0))
			elapsedTime = float32(time.Since(frameStart).Seconds())
		}

	}
}
