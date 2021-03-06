// Copyright 2017 The bluemun Authors. All rights reserved.
// Use of this source code is governed by a MIT License
// license that can be found in the LICENSE file.

// Package render renderer.go Defines a renderer in the graphics package.
package render

import (
	"github.com/bluemun/munfall"
	"github.com/bluemun/munfall/graphics/shader"
	"github.com/go-gl/gl/v3.3-core/gl"
)

// Renderer interface used to talk to renderers.
type Renderer interface {
	Begin()
	DrawRectangle(x, y, w, h float32, color uint32)
	Submit(ra munfall.Renderable)
	Flush()
	End()
}

// Renderer used by the graphics library to draw.
type renderer2d struct {
	vertexOffset, indexOffset, vertexBufferSize, indexBufferSize int
	vertexArray, vertexBuffer, indexBuffer                       uint32
	s                                                            *shader.Shader
}

const int32Size = 4
const float32Size = 4
const vertexSize = 4

const vertexShader = `
#version 130
in highp vec3 vertex;
in highp float color;

uniform highp mat4 pr;

out highp float fcolor;

void main() {
  fcolor = color;
  gl_Position = pr * vec4(vertex, 1);
}
` + "\x00"

const fragmentShader = `
#version 130
in highp float fcolor;
out vec4 outputColor;
void main() {
  uint c = uint(fcolor);
  outputColor = vec4(
    float((c >> 24u) & 0xffu) / 255.0,
    float((c >> 16u) & 0xffu) / 255.0,
    float((c >> 8u) & 0xffu) / 255.0,
    float(c & 0xffu) / 255.0
  );
}
` + "\x00"

// CreateRenderer2D used to create a renderer2d object correctly.
func CreateRenderer2D(vertexBufferSize, indexBufferSize int) Renderer {
	r := &renderer2d{
		vertexBufferSize: vertexBufferSize,
		indexBufferSize:  indexBufferSize,
	}

	munfall.Do(func() {
		r.s = shader.CreateShader(vertexShader, fragmentShader)
		r.s.Use()
		gl.GenVertexArrays(1, &r.vertexArray)
		munfall.CheckGLError()
		gl.BindVertexArray(r.vertexArray)
		munfall.CheckGLError()

		gl.GenBuffers(1, &r.vertexBuffer)
		munfall.CheckGLError()
		gl.BindBuffer(gl.ARRAY_BUFFER, r.vertexBuffer)
		munfall.CheckGLError()
		gl.BufferData(gl.ARRAY_BUFFER, (int)(10000*vertexSize*float32Size), nil, gl.DYNAMIC_DRAW)
		munfall.CheckGLError()

		point := r.s.GetAttributeLocation("vertex")
		gl.EnableVertexAttribArray(point)
		munfall.CheckGLError()
		munfall.Logger.Info("Vertex attribute vertex location: ", point)
		gl.VertexAttribPointer(point, 3, gl.FLOAT, false, vertexSize*float32Size, gl.PtrOffset(0))
		munfall.CheckGLError()

		color := r.s.GetAttributeLocation("color")
		gl.EnableVertexAttribArray(color)
		munfall.CheckGLError()
		munfall.Logger.Info("Vertex attribute color location: ", color)
		gl.VertexAttribPointer(color, 1, gl.FLOAT, false, vertexSize*float32Size, gl.PtrOffset(3*float32Size))
		munfall.CheckGLError()

		gl.GenBuffers(1, &r.indexBuffer)
		munfall.CheckGLError()
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.indexBuffer)
		munfall.CheckGLError()

		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, (int)(10000*int32Size), nil, gl.DYNAMIC_DRAW)
		munfall.CheckGLError()

		r.s.BindFragDataLocation("outputColor")
		munfall.CheckGLError()
	})

	return r
}

// Begin starts the rendering procedure.
func (r *renderer2d) Begin() {
	r.indexOffset, r.vertexOffset = 0, 0
	munfall.Do(func() {
		r.s.Use()
		if activeCamera != nil {
			activeCamera.use(r.s)
		}

		gl.BindBuffer(gl.ARRAY_BUFFER, r.vertexBuffer)
		munfall.CheckGLError()
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.indexBuffer)
		munfall.CheckGLError()
	})
}

// DrawRectangle draws a rectangle using the given values, x and y point to the top-left corner.
func (r *renderer2d) DrawRectangle(x, y, w, h float32, color uint32) {
	vertices := [16]float32{
		x, y, 0, float32(color),
		x + w, y, 0, float32(color),
		x, y + h, 0, float32(color),
		x + w, y + h, 0, float32(color),
	}
	indices := [6]uint32{
		0 + uint32(r.vertexOffset), 1 + uint32(r.vertexOffset), 2 + uint32(r.vertexOffset),
		1 + uint32(r.vertexOffset), 2 + uint32(r.vertexOffset), 3 + uint32(r.vertexOffset),
	}

	r.draw(vertices[:], indices[:])
}

// Submit adds the given Renderable to this draw call.
func (r *renderer2d) Submit(ra munfall.Renderable) {
	mesh := ra.Mesh()
	pos := ra.Pos()
	color := float32(ra.Color())

	var vertices []float32
	for i := 0; i < len(mesh.Points); i += 3 {
		vertices = append(vertices, mesh.Points[i]+pos.X, mesh.Points[i+1]+pos.Y, mesh.Points[i+2], color)
	}
	var indices []uint32
	for i := 0; i < len(mesh.Triangles); i++ {
		indices = append(indices, uint32(r.vertexOffset)+mesh.Triangles[i])
	}

	r.draw(vertices, indices)
}

func (r *renderer2d) draw(vertices []float32, indices []uint32) {
	if len(vertices) == 0 || len(indices) == 0 {
		return
	}

	if r.vertexOffset*vertexSize+len(vertices) >= r.vertexBufferSize || r.indexOffset+len(indices) >= r.indexBufferSize {
		r.Flush()
		r.indexOffset, r.vertexOffset = 0, 0
	}

	munfall.Do(func() {
		gl.BufferSubData(gl.ARRAY_BUFFER, (r.vertexOffset*vertexSize)*float32Size, len(vertices)*float32Size, gl.Ptr(vertices))
		munfall.CheckGLError()
		gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, (r.indexOffset)*int32Size, len(indices)*int32Size, gl.Ptr(indices))
		munfall.CheckGLError()
	})
	r.vertexOffset += len(vertices) / vertexSize
	r.indexOffset += len(indices)
}

// Flush flushes all the draw calls that have been called on this renderer to the window
func (r *renderer2d) Flush() {
	munfall.Do(func() {
		gl.BindVertexArray(r.vertexArray)
		munfall.CheckGLError()
		gl.DrawElements(gl.TRIANGLES, int32(r.indexOffset), gl.UNSIGNED_INT, nil)
		munfall.CheckGLError()
		gl.BindVertexArray(0)
		munfall.CheckGLError()
	})
}

// End ends the rendering procedure.
func (r *renderer2d) End() {
	munfall.Do(func() {
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
		munfall.CheckGLError()
		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		munfall.CheckGLError()
	})
}
