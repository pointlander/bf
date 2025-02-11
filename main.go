// Copyright 2025 The BF Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
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
			count := int(math.Abs(rng.NormFloat64()*128)) + 1
			for j := 0; j < count; j++ {
				program.WriteRune('+')
			}
		case 1:
			count := int(math.Abs(rng.NormFloat64()*128)) + 1
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

func main() {
	rng := rand.New(rand.NewSource(1))
	program := strings.Builder{}
	Generate(0, 2, rng, &program)
	fmt.Println(program.String())
	fmt.Println()
	p := Program(program.String())
	output := p.Execute(rng, 33)
	fmt.Println(output.String())
}
