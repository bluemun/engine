// Copyright 2017 The bluemun Authors. All rights reserved.
// Use of this source code is governed by a MIT License
// license that can be found in the LICENSE file.

// Package logic world.go Defines our world type that runs the game.
package logic

import (
	"github.com/bluemun/engine"
	"github.com/bluemun/engine/traits"
)

// World container that manages the game world.
type world struct {
	actors          map[uint]*actor
	traitDictionary *traitDictionary
	nextActorID     uint
	endtasks        []func()
}

// CreateWorld creates and initializes the World.
func CreateWorld() engine.World {
	world := &world{actors: make(map[uint]*actor, 10), endtasks: nil}
	world.traitDictionary = createTraitDictionary(world)
	return (engine.World)(world)
}

// AddFrameEndTask adds a task that will be run at the end of the current tick.
func (w *world) AddFrameEndTask(f func()) {
	w.endtasks = append(w.endtasks, f)
}

func (w *world) GetTrait(a engine.Actor, i interface{}) engine.Trait {
	return w.traitDictionary.GetTrait(a.(*actor), i)
}

func (w *world) GetTraitsImplementing(a engine.Actor, i interface{}) []engine.Trait {
	return w.traitDictionary.GetTraitsImplementing(a.(*actor), i)
}

func (w *world) GetAllTraitsImplementing(i interface{}) []engine.Trait {
	return w.traitDictionary.GetAllTraitsImplementing(i)
}

// RemoveActor removes the given actor from the world.
func (w *world) RemoveActor(a engine.Actor) {
	if a == nil {
		panic("Trying to remove nil as an Actor!")
	}

	notify := w.traitDictionary.GetTraitsImplementing(a.(*actor), (*traits.TraitRemovedNotifier)(nil))
	w.traitDictionary.removeActor(a.(*actor))
	delete(w.actors, a.GetActorID())
	for _, trait := range notify {
		trait.(traits.TraitRemovedNotifier).NotifyRemoved(a)
	}
}

// ResolveOrder bla.
func (w *world) ResolveOrder(order *engine.Order) {
	resolvers := w.traitDictionary.GetAllTraitsImplementing((*traits.TraitOrderResolver)(nil))
	for _, trait := range resolvers {
		trait.(traits.TraitOrderResolver).ResolveOrder(order)
	}
}

// Tick ticks all traits on the traitmanager that implement the Tick interface.
func (w *world) Tick(deltaUnit float32) {
	tickers := w.traitDictionary.GetAllTraitsImplementing((*traits.TraitTicker)(nil))
	for _, ticker := range tickers {
		ticker.(traits.TraitTicker).Tick(deltaUnit)
	}

	for _, task := range w.endtasks {
		task()
	}

	w.endtasks = nil
}
