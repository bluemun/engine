// Copyright 2017 The bluemun Authors. All rights reserved.
// Use of this source code is governed by a MIT License
// license that can be found in the LICENSE file.

// Package traits base.go Defines interfaces that define traits.
package traits

import (
	"github.com/bluemun/munfall"
)

// TraitTicker is a trait that gets called every time the world ticks.
type TraitTicker interface {
	munfall.Trait
	Tick(deltaUnit float32)
}

// TraitOrderResolver used by traits to resolve orders sent by an order generator.
type TraitOrderResolver interface {
	munfall.Trait
	ResolveOrder(order *munfall.Order)
}
