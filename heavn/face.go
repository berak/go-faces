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
func rectbit(m *Matrix, pos int, cen uint8) (uint8) {
    way := getbit(m, pos, cen)
    way += getbit(m, pos+1, cen)
    way += getbit(m, pos+m.w, cen)
    way += getbit(m, pos+m.w+1, cen)
    if way >= 2 { return 1 }
    return 0
}

// Sampler (olbp)
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

// Sampler (hardcoded 8,2 elbp)
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

// Sampler (olbp variant that interpolates for all points, nut just the corners)
func square2(m *Matrix, pos,radius int) (way uint8) {
    cen := m.e[pos]
    way = 0
    way |= rectbit(m, pos-radius*m.w,              cen) << 0
    way |= rectbit(m, pos-radius*m.w+radius-1,     cen) << 1
    way |= rectbit(m, pos+radius-1,                cen) << 2
    way |= rectbit(m, pos+(radius-1)*m.w+radius-1, cen) << 3
    way |= rectbit(m, pos+(radius-1)*m.w,          cen) << 4
    way |= rectbit(m, pos+(radius-1)*m.w-radius,   cen) << 5
    way |= rectbit(m, pos-radius,                  cen) << 6
    way |= rectbit(m, pos-radius*m.w-radius,       cen) << 7
    return
}

// Sampler  (variable radius, 8 neighbours) elbp
func circle2(m *Matrix, pos,radius int) (way uint8) {
    cen := m.e[pos]
    way = 0
    way |= getbit( m, pos-radius*m.w,                cen) << 0
    way |= rectbit(m, pos-radius*m.w+(radius-1),     cen) << 1
    way |= getbit( m, pos+radius,                    cen) << 2
    way |= rectbit(m, pos+(radius-1)+(radius-1)*m.w, cen) << 3
    way |= getbit( m, pos+radius*m.w,                cen) << 4
    way |= rectbit(m, pos-(radius)-(radius-1)*m.w,   cen) << 5
    way |= getbit( m, pos-radius,                    cen) << 6
    way |= rectbit(m, pos-radius-radius*m.w,         cen) << 7
    return
}

// as reference, this is similar to the opencv one
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
// equalize luminance with normalized histogram
//
func (m *Matrix) equalize_hist() {
    hist_sz := 256
    img_sz  := m.h * m.w
    hist    := make(Histogram, hist_sz)
    
    for i:=0; i<img_sz; i++ {
        hist[ m.e[i] ] ++
    }

    scale := 255.0 / float64(img_sz)
    lut   := make([]uint8, hist_sz)
    sum   := float64(0)
    
    for i:=0; i<hist_sz; i++ {
        sum += float64(hist[i])
        val := math.Floor(0.5 + sum*scale)
        //if val > 255.0 { val = 255.0 }
        lut[i] = uint8( val )
    }

    lut[0] = 0
    for i:=0; i<img_sz; i++ {
        m.e[i] = lut[m.e[i]]
    }
}

//
// unified spatial(7x7) histogram with preprocessed indices
//
func (m *Matrix) histogram(sample Sampler, radius int) *Histogram {
    ncell := 7
    nhist := 57  // + 1  aww, i lost one (actually, 2)!
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

func chi_square(a, b *Histogram ) float64 {
    if len(*a) != len(*b) {return -1}
    dist := 0.0
    for ai, av := range *a {
        sum := float64(av) + float64((*b)[ai])
        dif := float64(av) - float64((*b)[ai])
        if (sum>0) {
            dist += dif*dif/sum
        }
    }
    return dist
}

func chi_square2(a, b *Histogram ) float64 {
    if len(*a) != len(*b) {return -1}
    dist := 0.0
    for ai, av := range *a {
        sum :=               float64((*b)[ai])
        dif := float64(av) - float64((*b)[ai])
        if (sum>0) {
            dist += dif*dif/sum
        }
    }
    return dist
}


//
// preprocess the indexing [see init()]:
// in the uniform u8 case, there are only 58 of 255 valid cases
//
func bit(b, i uint8) bool {
    return ((b & (1 << i)) != 0)
}

// walk the bitmask and count ups/downs
func uniform(way uint8) bool {
    cu := 0
    if bit(way, 0) != bit(way, 1) {    cu += 1    }
    if bit(way, 1) != bit(way, 2) {    cu += 1    }
    if bit(way, 2) != bit(way, 3) {    cu += 1    }
    if bit(way, 3) != bit(way, 4) {    cu += 1    }
    if bit(way, 4) != bit(way, 5) {    cu += 1    }
    if bit(way, 5) != bit(way, 6) {    cu += 1    }
    if bit(way, 6) != bit(way, 7) {    cu += 1    }
    if bit(way, 7) != bit(way, 0) {    cu += 1    }
    return (cu <= 2)
}

func initIndices() {
    iu := 0
    unindices = make([]int, 256)
    for i := uint8(0); i < 0xff; i++ {
        if uniform(i) {
            unindices[i] = iu
            iu += 1
        } else {
            unindices[i] = -1
        }
    }
}

// [0 1 2 3 4 -1 5 6 7 -1 -1 -1 8 -1 9 10 11 -1 -1 -1 -1 -1 -1 -1 12 -1 -1 -1 13 -1 14 15 16 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 17 -1 -1 -1 -1 -1 -1 -1 18 -1 -1 -1 19 -1 20 21 22 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 23 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 24 -1 -1 -1 -1 -1 -1 -1 25 -1 -1 -1 26 -1 27 28 29 30 -1 31 -1 -1 -1 32 -1 -1 -1 -1 -1 -1 -1 33 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 34 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 35 36 37 -1 38 -1 -1 -1 39 -1 -1 -1 -1 -1 -1 -1 40 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 -1 41 42 43 -1 44 -1 -1 -1 45 -1 -1 -1 -1 -1 -1 -1 46 47 48 -1 49 -1 -1 -1 50 51 52 -1 53 54 55 56 57]
