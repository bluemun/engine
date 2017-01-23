// Copyright 2017 The bluemun Authors. All rights reserved.
// Use of this source code is governed by a MIT License
// license that can be found in the LICENSE file.

// Package game game.go Defines the struct used to connect all the engine components together.
package game

import (
	"runtime"
	"time"

	"github.com/bluemun/engine"
	"github.com/bluemun/engine/graphics"
	"github.com/bluemun/engine/graphics/render"
	"github.com/bluemun/engine/input"
	"github.com/bluemun/engine/logic"
)

var mainHasRun = false

// Game type used to gold all the components needed to run a game.
type Game struct {
	Camera         *render.Camera
	orderGenerator *input.OrderGenerator
	window         *graphics.Window
	world          *logic.World
	renderer       render.RendersTraits
}

// Initialize initializes the game.
func (g *Game) Initialize() {
	if !mainHasRun {
		mainHasRun = true
		go func() {
			runtime.LockOSThread()
			engine.Loop()
		}()
	}

	g.window = graphics.CreateWindow()
	g.Camera = &render.Camera{}
	g.Camera.Activate()

	g.world = logic.CreateWorld()

	// TODO: Change this once we got more renderers.
	g.renderer = render.CreateRendersTraits2D(g.world)
}

// Start starts the game loop, doesn't return untill the game is closed.
func (g *Game) Start() {
	render := time.NewTicker(time.Second / 60)
	update := time.NewTicker(time.Second / 60)

	for {
		select {
		case <-render.C:
			g.window.Clear()
			g.renderer.Render()
			g.window.SwapBuffers()
		case <-update.C:
			g.world.Tick(1 / 60.0)
			g.window.PollEvents()
			if g.window.Closed() {
				render.Stop()
				update.Stop()
				close(engine.Mainfunc)
			}
		}
	}
}

// SetOrderGenerator sets the current active order generator for the game.
func (g *Game) SetOrderGenerator(og *input.OrderGenerator) {
	g.orderGenerator = og
}

// World returns the underlying world.
func (g *Game) World() *logic.World {
	return g.world
}
