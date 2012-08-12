package main

import (
	"math"
)

var unindices []int = nil


type Histogram []uint16


type PicRec struct {
	Name string
	Hist []byte
}

type Matrix struct {
	e    []uint8
	w, h int
}

type Sampler func(m *Matrix, pos,radius int) uint8



func createMatrix(w, h int) *Matrix {
	var m Matrix
	m.w = w
	m.h = h
	m.e = make([]uint8, w*h)
	return &m
}
func (m *Matrix) pos(x, y int) int {
	return (m.w*y + x)
}
func (m *Matrix) set(x, y int, val uint8) {
	m.e[m.w*y+x] = val
}
func (m *Matrix) get(x, y int) uint8 {
	return m.e[m.w*y+x]
}

func getbit(m *Matrix, pos int, cen uint8) uint8 {
	if m.e[pos] >= cen {
		return 1
	}
	return 0
}

// Sampler
func square(m *Matrix, pos,radius int) (way uint8) {
	cen := m.e[pos]
	way = 0
	way |= getbit(m, pos-radius*m.w,        cen) << 0
	way |= getbit(m, pos-radius*m.w+radius, cen) << 1
	way |= getbit(m, pos+radius,            cen) << 2
	way |= getbit(m, pos+radius*m.w+radius, cen) << 3
	way |= getbit(m, pos+radius*m.w,        cen) << 4
	way |= getbit(m, pos+radius*m.w-radius, cen) << 5
	way |= getbit(m, pos-radius,            cen) << 6
	way |= getbit(m, pos-radius*m.w-radius, cen) << 7
	return
}

// Sampler
func circle(m *Matrix, pos,radius int) (way uint8) {
	cen := m.e[pos]
	tmpway := uint8(0)
	way = 0
	way    |= getbit(m, pos-m.w, cen) << 0
	tmpway  = getbit(m, pos-m.w+1, cen)
	tmpway += getbit(m, pos+1, cen)
	tmpway += getbit(m, pos, cen)
	tmpway += getbit(m, pos-m.w, cen)
	if tmpway >= 2 {
		way |= 1 << 1
	}
	way    |= getbit(m, pos+1, cen) << 2
	tmpway  = getbit(m, pos+m.w+1, cen)
	tmpway += getbit(m, pos+m.w, cen)
	tmpway += getbit(m, pos, cen)
	tmpway += getbit(m, pos+1, cen)
	if tmpway >= 2 {
		way |= 1 << 3
	}
	way    |= getbit(m, pos+m.w, cen) << 4
	tmpway  = getbit(m, pos+m.w-1, cen)
	tmpway += getbit(m, pos-1, cen)
	tmpway += getbit(m, pos, cen)
	tmpway += getbit(m, pos+m.w, cen)
	if tmpway >= 2 {
		way |= 1 << 5
	}
	way    |= getbit(m, pos-1, cen) << 6
	tmpway  = getbit(m, pos-m.w-1, cen)
	tmpway += getbit(m, pos-m.w, cen)
	tmpway += getbit(m, pos, cen)
	tmpway += getbit(m, pos-1, cen)
	if tmpway >= 2 {
		way |= 1 << 7
	}
	return
}

func rectbit(m *Matrix, pos int, cen uint8) (uint8) {
	way := getbit(m, pos, cen)
	way += getbit(m, pos+1, cen)
	way += getbit(m, pos+m.w, cen)
	way += getbit(m, pos+m.w+1, cen)
	if way >= 2 { return 1 }
	return 0
}
// Sampler
func square2(m *Matrix, pos,radius int) (way uint8) {
	cen := m.e[pos]
	way = 0
	way |= rectbit(m, pos-radius*m.w,            cen) << 0
	way |= rectbit(m, pos-radius*m.w+radius-1,   cen) << 1
	way |= rectbit(m, pos+radius-1,              cen) << 2
	way |= rectbit(m, pos+(radius-1)*m.w+radius-1, cen) << 3
	way |= rectbit(m, pos+(radius-1)*m.w,        cen) << 4
	way |= rectbit(m, pos+(radius-1)*m.w-radius, cen) << 5
	way |= rectbit(m, pos-radius,                cen) << 6
	way |= rectbit(m, pos-radius*m.w-radius,     cen) << 7
	return
}
// Sampler
func circle2(m *Matrix, pos,radius int) (way uint8) {
	cen := m.e[pos]
	way = 0
	way |= getbit( m, pos-radius*m.w, 			    cen) << 0
	way |= rectbit(m, pos-radius*m.w+(radius-1),    cen) << 1
	way |= getbit( m, pos+radius, 			        cen) << 2
	way |= rectbit(m, pos+(radius-1)+(radius-1)*m.w,cen) << 3
	way |= getbit( m, pos+radius*m.w, 	            cen) << 4
	way |= rectbit(m, pos-(radius)-(radius-1)*m.w,  cen) << 5
	way |= getbit( m, pos-radius,			        cen) << 6
	way |= rectbit(m, pos-radius-radius*m.w,        cen) << 7
	return
}

// Sampler
func elbp(m *Matrix, pos,radius int) (way uint8) {
	way = 0
	cen := m.e[pos]
    for n:=uint8(0); n<8; n++ {
        x  := float64(-radius) * math.Sin(math.Pi*float64(n)/4.0);
        y  := float64( radius) * math.Cos(math.Pi*float64(n)/4.0);
        fx := math.Floor(x);
        fy := math.Floor(y);
        cx := math.Ceil(x);
        cy := math.Ceil(y);
        // fractional part
        ty := y - fy;
        tx := x - fx;
        // set interpolation weights
        w1 := (1.0 - tx) * (1.0 - ty);
        w2 :=        tx  * (1.0 - ty);
        w3 := (1.0 - tx) *        ty;
        w4 :=        tx  *        ty;
		t  := w1*float64(m.e[pos + int(fx) + int(fy)*m.w]);
		t  += w2*float64(m.e[pos + int(cx) + int(fy)*m.w]);
		t  += w3*float64(m.e[pos + int(fx) + int(cy)*m.w]);
		t  += w4*float64(m.e[pos + int(cx) + int(cy)*m.w]);
		if uint8(t) > cen {
			way |= 1<<n
		}
    }
	return
}

//
// unified spatial(7x7) histogram with preprocessed indices
//
func (m *Matrix) histogram(sample Sampler, radius int) *Histogram {
	ncell := 7
	nhist := 57
	cw  := int(m.w / (ncell-1))
	ch  := int(m.h / (ncell-1))
	hist := make(Histogram, nhist*(ncell*ncell))
	for y := radius; y < m.h-radius; y++ {
		for x := radius; x < m.w-radius; x++ {
			way := sample(m, m.pos(x, y),radius)
			idx := unindices[way]

			cx  := int(x / cw)
			cy  := int(y / ch)
			ch  := (cx + cy * ncell) * nhist

			if idx != -1 {
				hist[ch + idx] += 1
			} else {
				// assign non-uniform ones to the 'noise' channel
				hist[ch + nhist-1] += 1
			}
		}
	}
	return &hist
}

func norml2(a, b *Histogram) float64 {
	if len(*a) != len(*b) {return -1}
	dist := int(0)
	for ai, av := range *a {
		d := int(av) - int((*b)[ai])
		dist += d * d
	}
	return math.Sqrt(float64(dist))
}



//
// preprocess the indexing [see init()]:
// in the uniform u8 case, there are only 57 of 255 valid cases
//
func bit(b, i uint8) bool {
	return ((b & (1 << i)) != 0)
}
func uniform(way uint8) bool {
	cu := 0
	if bit(way, 0) != bit(way, 1) {	cu += 1	}
	if bit(way, 1) != bit(way, 2) {	cu += 1	}
	if bit(way, 2) != bit(way, 3) {	cu += 1	}
	if bit(way, 3) != bit(way, 4) {	cu += 1	}
	if bit(way, 4) != bit(way, 5) {	cu += 1	}
	if bit(way, 5) != bit(way, 6) {	cu += 1	}
	if bit(way, 6) != bit(way, 7) {	cu += 1	}
	if bit(way, 7) != bit(way, 0) {	cu += 1	}
	return (cu <= 2)
}
func initIndices() {
	iu := 0
	unindices = make([]int, 257)
	unindices[256] = -1
	for i := uint8(0); i < 0xff; i++ {
		if uniform(i) {
			unindices[i] = iu
			iu += 1
		} else {
			unindices[i] = -1
		}
	}
}




