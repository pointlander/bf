// Copyright 2025 The BF Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"math"
	"math/rand"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

const (
	// MemorySize is the size of the working memory
	MemorySize = 1024 * 1024
	// CyclesLimit is the limit on cycles
	CyclesLimit = 1024 * 1024
)

var (
	// Genes are the genes
	Genes = [...]rune{'+', '-', '>', '<', '.', '[', ']'}
)

// Program is a program
// https://github.com/cvhariharan/goBrainFuck
type Program []rune

// Execute executes a program
func (p Program) Execute(rng *rand.Rand, size int) *strings.Builder {
	var (
		memory [MemorySize]int
		pc     int
		dc     int
		i      int
		output strings.Builder
	)
	length := len(p)

	for pc < length && i < CyclesLimit {
		opcode := p[pc]
		switch opcode {
		case '+':
			memory[dc%MemorySize] += 1
			pc++
		case '-':
			memory[dc%MemorySize] -= 1
			pc++
		case '>':
			dc++
			pc++
		case '<':
			if dc > 0 {
				dc--
			}
			pc++
		case '.':
			m := memory[dc%MemorySize]
			if m < 0 {
				m = -m
			}
			output.WriteRune(Genes[m%len(Genes)])
			if len([]rune(output.String())) == size {
				return &output
			}
			pc++
		case ',':
			memory[dc] = rng.Intn(len(Genes))
			pc++
		case '[':
			if memory[dc] == 0 {
				pc = p.findMatchingForward(pc) + 1
			} else {
				pc++
			}
		case ']':
			if memory[dc] != 0 {
				pc = p.findMatchingBackward(pc) + 1
			} else {
				pc++
			}
		default:
			pc++
		}
		i++
	}
	return &output
}

func (p Program) findMatchingForward(position int) int {
	count, length := 1, len(p)
	for i := position + 1; i < length; i++ {
		if p[i] == ']' {
			count--
			if count == 0 {
				return i
			}
		} else if p[i] == '[' {
			count++
		}
	}

	return length - 1
}

func (p Program) findMatchingBackward(position int) int {
	count := 1
	for i := position - 1; i >= 0; i-- {
		if p[i] == '[' {
			count--
			if count == 0 {
				return i
			}
		} else if p[i] == ']' {
			count++
		}
	}

	return -1
}

func Generate(depth, limit int, rng *rand.Rand, program *strings.Builder) {
	if depth > limit {
		return
	}
	count := rng.Intn(128) + 1
	for i := 0; i < count; i++ {
		switch rng.Intn(10) {
		case 0:
			count := int(math.Abs(rng.NormFloat64()*16)) + 1
			for j := 0; j < count; j++ {
				program.WriteRune('+')
			}
		case 1:
			count := int(math.Abs(rng.NormFloat64()*16)) + 1
			for j := 0; j < count; j++ {
				program.WriteRune('-')
			}
		case 2, 3:
			program.WriteRune('>')
		case 4, 5:
			program.WriteRune('<')
		case 6, 7:
			program.WriteRune('.')
		case 8, 9:
			program.WriteRune('[')
			Generate(depth+1, limit, rng, program)
			program.WriteRune(']')
		}
	}
}

// Vector is a vector
type Vector struct {
	Vector [256]float32
	Symbol byte
}

func main() {
	rng := rand.New(rand.NewSource(1))
	/*program := strings.Builder{}
	Generate(0, 2, rng, &program)
	fmt.Println(program.String())
	fmt.Println()
	p := Program(program.String())
	output := p.Execute(rng, 33)
	fmt.Println(output.String())*/

	rows := make(plotter.XYs, 0, 8)
	heads := make(plotter.XYs, 0, 8)
	values := make(plotter.XYs, 0, 8)
	m, tape, head, pool, index := NewMixer(), [1024]byte{}, 0, [1024]Vector{}, 0
	for i := range pool {
		for j := range pool[i].Vector {
			pool[i].Vector[j] = rng.Float32()
		}
		pool[i].Symbol = byte(rng.Intn(256))
	}
	m.Add(0)
	for i := 0; i < 4*4094; i++ {
		output := m.Mix()
		max, v, row := float32(0.0), 0, 0
		for r := 0; r < output.Rows; r++ {
			for key, value := range pool {
				cs := CS(output.Data[r*output.Cols:(r+1)*output.Cols], value.Vector[:])
				if cs > max {
					max, v, row = cs, key, r
				}
			}
		}
		if pool[v].Symbol&1 == 0 {
			head = (head + 1) % len(tape)
		} else {
			head = (head - 1 + len(tape)) % len(tape)
		}
		tape[head] = tape[head] ^ pool[v].Symbol
		index = (index + 1) % len(pool)
		pool[index].Symbol = tape[head]
		rows = append(rows, plotter.XY{X: float64(i), Y: float64(row)})
		heads = append(heads, plotter.XY{X: float64(i), Y: float64(head)})
		values = append(values, plotter.XY{X: float64(i), Y: float64(tape[head])})
		m.Add(tape[head])
	}

	{
		p := plot.New()

		p.Title.Text = "row vs time"
		p.X.Label.Text = "time"
		p.Y.Label.Text = "row"

		scatter, err := plotter.NewScatter(rows)
		if err != nil {
			panic(err)
		}
		scatter.GlyphStyle.Radius = vg.Length(1)
		scatter.GlyphStyle.Shape = draw.CircleGlyph{}
		p.Add(scatter)

		err = p.Save(8*vg.Inch, 8*vg.Inch, "rows.png")
		if err != nil {
			panic(err)
		}
	}

	{
		p := plot.New()

		p.Title.Text = "head vs time"
		p.X.Label.Text = "time"
		p.Y.Label.Text = "head"

		scatter, err := plotter.NewScatter(heads)
		if err != nil {
			panic(err)
		}
		scatter.GlyphStyle.Radius = vg.Length(1)
		scatter.GlyphStyle.Shape = draw.CircleGlyph{}
		p.Add(scatter)

		err = p.Save(8*vg.Inch, 8*vg.Inch, "heads.png")
		if err != nil {
			panic(err)
		}
	}

	{
		p := plot.New()

		p.Title.Text = "value vs time"
		p.X.Label.Text = "time"
		p.Y.Label.Text = "value"

		scatter, err := plotter.NewScatter(values)
		if err != nil {
			panic(err)
		}
		scatter.GlyphStyle.Radius = vg.Length(1)
		scatter.GlyphStyle.Shape = draw.CircleGlyph{}
		p.Add(scatter)

		err = p.Save(8*vg.Inch, 8*vg.Inch, "values.png")
		if err != nil {
			panic(err)
		}
	}

}
