package game

import "math/rand"
import "bytes"

type Field struct {
	s    [][]bool
	w, h int
}

func NewField(w, h int) *Field {
	s := make([][]bool, h)
	for i := range s {
		s[i] = make([]bool, w)
	}
	return &Field{s: s, w: w, h: h}
}

func (f *Field) Set(x, y int, b bool) {
	f.s[y][x] = b
}

func (f *Field) Get(x, y int) bool {
	return f.s[y][x]
}

// If the x or y coordinates are outside the field boundaries they are wrapped
// toroidally. For instance, an x value of -1 is treated as width-1.
func (f *Field) Alive(x, y int) bool {
	x += f.w
	x %= f.w
	y += f.h
	y %= f.h
	return f.s[y][x]
}

func (f *Field) Next(x, y int) bool {
	// Count the adjacent cells that are alive.
	alive := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if (j != 0 || i != 0) && f.Alive(x+i, y+j) {
				alive++
			}
		}
	}
	// Return next state according to the game rules:
	//   exactly 3 neighbors: on,
	//   exactly 2 neighbors: maintain current state,
	//   otherwise: off.
	return alive == 3 || alive == 2 && f.Alive(x, y)
}

type Life struct {
	a, b *Field
	w, h int
}

func NewLife(w, h int) *Life {
	a := NewField(w, h)
	return &Life{
		a: a, b: NewField(w, h),
		w: w, h: h,
	}
}

func (l *Life) RandomInit() {
	for i := 0; i < (l.w * l.h / 4); i++ {
		l.a.Set(rand.Intn(l.w), rand.Intn(l.h), true)
	}
}

func (l *Life) Clear() {
	for y := 0; y < l.h; y++ {
		for x := 0; x < l.w; x++ {
			l.a.Set(x, y, false)
		}
	}
}

func (l *Life) W() int {
	return l.w;
}

func (l *Life) H() int {
	return l.h;
}

func (l *Life) Alive(x, y int) bool {
	return l.a.Alive(x, y)
}

func (l *Life) Toggle(x, y int) {
	if l.a.Get(x, y) {
		l.a.Set(x, y, false)
	} else {	 	
		l.a.Set(x, y, true)
	}
}

func (l *Life) Step() {
	// Update the state of the next field (b) from the current field (a).
	for y := 0; y < l.h; y++ {
		for x := 0; x < l.w; x++ {
			l.b.Set(x, y, l.a.Next(x, y))
		}
	}
	// Swap fields a and b.
	l.a, l.b = l.b, l.a
}

func (l *Life) String() string {
	var buf bytes.Buffer
	for y := 0; y < l.h; y++ {
		for x := 0; x < l.w; x++ {
			b := byte(' ')
			if l.a.Alive(x, y) {
				b = '*'
			}
			buf.WriteByte(b)
		}
		buf.WriteByte('\n')
	}
	return buf.String()
}