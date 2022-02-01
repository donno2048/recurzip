package main
import (
	"bufio"
	"bytes"
	"compress/flate"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"strconv"
	"errors"
)
func main() {
	makeGz()
	// makeZip()
}
func makeGz() {
	head := []byte{
		0x1f, 0x8b, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		'r', 'e', 'c', 'u', 'r', 's', 'i', 'v', 'e', 0x00,
	}
	zhead := deflate(head, true, false)
	ztail := make([]byte, 5+8)
	ztail[0] = 1
	ztail[1] = 8
	ztail[2] = 0
	ztail[3] = ^byte(8)
	ztail[4] = ^byte(0)
	tail := ztail[5:]
	tail[0] = 0xaa
	tail[1] = 0xbb
	tail[2] = 0xcc
	tail[3] = 0xdd
	_, whole := makeGeneric(zhead, head, ztail, tail, nil)
	n := len(whole)
	tail[4] = byte(n)
	tail[5] = byte(n>>8)
	tail[6] = byte(n>>16)
	tail[7] = byte(n>>24)
	_, whole = makeGeneric(zhead, head, ztail, tail, tail[0:4])
	f, _ := os.OpenFile("recursive.gz", os.O_CREATE|os.O_WRONLY, 0666)
	f.Write(whole)
	f.Close()
}
func makeZip() {
	csize := 0
	uncsize := 0
	sufpos := 0
	zhead := []byte{
		0x00, 37, 0, ^byte(37), 0xFF,	// 37-byte literal
		0x50, 0x4b, 0x03, 0x04,
		0x14,
		0x00,
		0x00, 0x00,
		0x08, 0x00,
		0x08, 0x03, // modtime
		0x64, 0x3c,	// moddate
		0xaa, 0xbb, 0xcc, 0xdd,
		byte(csize), byte(csize>>8), 0, 0,
		byte(uncsize), byte(uncsize>>8), 0, 0,
		0x07, 0x00,
		0x00, 0x00,
		'r', '/', 'r', '.', 'z', 'i', 'p',	// file name
	}
	head := zhead[5:]
	headsize := head[14:26]
	tail := []byte{
		0x50, 0x4b, 0x01, 0x02,
		0x14,
		0x00,
		0x14,
		0x00,
		0x00, 0x00,
		0x08, 0x00,
		0x08, 0x03,	// modtime
		0x64, 0x3c,	// moddate
		0xaa, 0xbb, 0xcc, 0xdd,
		byte(csize), byte(csize>>8), 0, 0,
		byte(uncsize), byte(uncsize>>8), 0, 0,
		0x07, 0x00,
		0x00, 0x00,
		0x00, 0x00,
		0x00, 0x00,
		0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		'r', '/', 'r', '.', 'z', 'i', 'p',	// file name
		0x50, 0x4b, 0x05, 0x06,
		0x00, 0x00,
		0x00, 0x00,
		0x01, 0x00,
		0x01, 0x00,
		53, 0x00,	0x00, 0x00, // size
		byte(sufpos), byte(sufpos>>8), 0x00, 0x00,
		0x00, 0x00,
	}
	var b wbuf
	var zero [12]byte
	b.writeBits(0, 1, false)
	b.writeBits(1, 2, false)
	b.writeBits(0x50+48, 8, true)
	b.writeBits(0x4b+48, 8, true)
	b.writeBits(0x01+48, 8, true)
	b.writeBits(0x02+48, 8, true)
	b.writeBits(0x14+48, 8, true)
	b.writeBits(0x00+48, 8, true)
	b.writeBits(270-256, 7, true)
	b.writeBits(1, 2, false)
	b.writeBits(16, 5, true)
	b.writeBits(367-256-1, 7, false)
	b.writeBits(267-256, 7, true)	
	b.writeBits(1, 1, false)
	b.writeBits(0, 5, true)
	b.writeBits('r'+48, 8, true)	// file name
	b.writeBits('/'+48, 8, true)
	b.writeBits('r'+48, 8, true)
	b.writeBits('.'+48, 8, true)
	b.writeBits('z'+48, 8, true)
	b.writeBits('i'+48, 8, true)
	b.writeBits('p'+48, 8, true)
	b.writeBits(0x50+48, 8, true)
	b.writeBits(0x4b+48, 8, true)
	b.writeBits(0x05+48, 8, true)
	b.writeBits(0x06+48, 8, true)
	b.writeBits(4-2, 7, true)
	b.writeBits(7, 5, true)
	b.writeBits(3, 2, false)
	b.writeBits(0x01+48, 8, true)
	b.writeBits(3-2, 7, true)
	b.writeBits(2-1, 5, true)
	b.writeBits(53+48, 8, true)	// size
	b.writeBits(0x00+48, 8, true)
	b.writeBits(0x00+48, 8, true)
	b.writeBits(0x00+48, 8, true)
	b.writeBits(0, 7, true)
	b.writeBits(1, 1, false)
	b.writeBits(0, 2, false)
	b.flushBits()
	b.bytes.WriteByte(6)
	b.bytes.WriteByte(0)
	b.bytes.WriteByte(^byte(6))
	b.bytes.WriteByte(^byte(0))
	tailsufOffset := b.bytes.Len()
	b.bytes.Write(zero[0:6])
	ztail := b.bytes.Bytes()
	tailsuf := ztail[tailsufOffset:tailsufOffset+4]
	_, whole := makeGeneric(zhead, head, ztail, tail, nil)
	csize = len(whole) - len(head) - len(tail)
	uncsize = len(whole)
	headsize[4+0] = byte(csize) 
	headsize[4+1] = byte(csize>>8)
	headsize[8+0] = byte(uncsize)
	headsize[8+1] = byte(uncsize>>8)
	tail[20] = byte(csize)
	tail[21] = byte(csize>>8)
	tail[24] = byte(uncsize)
	tail[25] = byte(uncsize>>8)
	sufpos = len(head) + csize
	tailsuf[0+0] = byte(sufpos)
	tailsuf[0+1] = byte(sufpos>>8)
	tail[len(tail)-6+0] = byte(sufpos)
	tail[len(tail)-6+1] = byte(sufpos>>8)
	_, whole = makeGeneric(zhead, head, ztail, tail, headsize[0:4])
	f, _ := os.OpenFile("r.zip", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	f.Write(whole)
	f.Close()
}
func makeGeneric(zhead, head, ztail, tail, crc []byte) (z, whole []byte) {
	const unit = 5
	var b wbuf
	b.bytes.Write(zhead)
	b.lit(len(zhead)+unit)
	b.bytes.Write(zhead)
	b.lit(len(zhead)+unit)
	b.rep(len(zhead)+unit)
	b.lit(unit)
	b.rep(len(zhead)+unit)
	b.lit(unit)
	b.lit(unit)
	b.lit(4*unit)
	b.rep(len(zhead)+unit)
	b.lit(unit)
	b.lit(unit)
	b.lit(4*unit)
	b.rep(4*unit)
	b.lit(4*unit)
	b.rep(4*unit)
	b.lit(4*unit)
	b.rep(4*unit)
	b.lit(4*unit)
	b.rep(4*unit)
	b.lit(4*unit)
	b.rep(4*unit)
	b.lit(0)
	b.lit(0)
	b.lit(len(ztail)+2*unit)
	b.rep(4*unit)
	b.lit(0)
	b.lit(0)
	b.lit(len(ztail)+2*unit)
	b.rep(len(ztail)+2*unit)
	b.lit(0)
	b.bytes.Write(ztail)
	b.rep(len(ztail)+2*unit)
	b.lit(0)
	b.bytes.Write(ztail)
	out := b.bytes.Bytes()
	{
		r := NewInflater(bytes.NewBuffer(out))
		var b1 bytes.Buffer
		io.Copy(&b1, r)
		r.Close()
		var b2 bytes.Buffer
		b2.Write(head)
		b2.Write(out)
		b2.Write(tail)
		whole = b1.Bytes()
	}
	if crc != nil {
		n := bytes.Count(whole, crc)
		embed := make([]int, n)
		off := 0
		for i := 0; i < n; i++ {
			j := bytes.Index(whole[off:], crc)
			off += j
			embed[i] = off
			off += 4
		}
		crc0 := uint32(0)
		crcbase := crc32.ChecksumIEEE(whole[0:embed[0]])
		for {
			if crc0&0xfffff == 0 {
				fmt.Printf("%#f%%\r", 100 * float64(crc0) / 0xffffffff)
			}
			for _, i := range embed {
				whole[i+0] = byte(crc0)
				whole[i+1] = byte(crc0>>8)
				whole[i+2] = byte(crc0>>16)
				whole[i+3] = byte(crc0>>24)
			}
			crc1 := crc32.Update(crcbase, crc32.IEEETable, whole[embed[0]:])
			if crc0 == crc1 {
				break
			}
			crc0++
		}
		fmt.Printf("SUCCESS   \n")
	}
	{
		r := NewInflater(bytes.NewBuffer(whole[len(head):len(head)+len(out)]))
		var b1 bytes.Buffer
		io.Copy(&b1, r)
		r.Close()
		whole = b1.Bytes()
	}
	return out, whole
}
type wbuf struct {
	bytes bytes.Buffer
	bit uint32
	nbit uint
	final uint32
}
func (b *wbuf) writeBits(bit uint32, nbit uint, rev bool) {
	if rev {
		br := uint32(0)
		for i := uint(0); i < nbit; i++ {
			if bit&(1<<i) != 0 {
				br |= 1<<(nbit-1-i)
			}
		}
		bit = br
	}
	b.bit |= bit << b.nbit
	b.nbit += nbit
	for b.nbit >= 8 {
		b.bytes.WriteByte(byte(b.bit))
		b.bit >>= 8
		b.nbit -= 8
	}
}
func (b *wbuf) flushBits() {
	if b.nbit > 0 {
		b.bytes.WriteByte(byte(b.bit))
		b.nbit = 0
		b.bit = 0
	}
}
func (b *wbuf) lit(n int) {
	b.writeBits(b.final, 1, false)
	b.writeBits(0, 2, false)
	b.flushBits()
	b1 := byte(n)
	b2 := byte(n>>8)
	b.bytes.WriteByte(b1)
	b.bytes.WriteByte(b2)
	b.bytes.WriteByte(^b1)
	b.bytes.WriteByte(^b2)
}
func (b *wbuf) rep(n int) {
	b.writeBits(b.final, 1, false)
	b.writeBits(1, 2, false)
	steal := uint(0)
	switch {
	case 9 <= n && n <= 12:
		b.writeBits(uint32(254+n/2)-256, 7, true)
		b.writeBits(6, 5, true)
		b.writeBits(uint32(n-8-1), 2, false)
		b.writeBits(uint32(254+n-n/2)-256, 7, true)
		b.writeBits(6, 5, true)
		b.writeBits(uint32(n-8-1), 2, false)
	case 13 <= n && n <= 16:
		b.writeBits(uint32(254+n/2)-256, 7, true)
		b.writeBits(7, 5, true)
		b.writeBits(uint32(n-12-1), 2, false)
		b.writeBits(uint32(254+n-n/2)-256, 7, true)
		b.writeBits(7, 5, true)
		b.writeBits(uint32(n-12-1), 2, false)
	case 17 <= n && n <= 20:
		b.writeBits(uint32(254+n/2)-256, 7, true)
		b.writeBits(8, 5, true)
		b.writeBits(uint32(n-16-1), 3, false)
		b.writeBits(uint32(254+n-n/2)-256, 7, true)
		b.writeBits(8, 5, true)
		b.writeBits(uint32(n-16-1), 3, false)
	case n == 21:
		b.writeBits(uint32(254+10)-256, 7, true)
		b.writeBits(8, 5, true)
		b.writeBits(uint32(n-16-1), 3, false)
		b.writeBits(uint32(265)-256, 7, true)
		b.writeBits(0, 1, true)
		b.writeBits(8, 5, true)
		b.writeBits(uint32(n-16-1), 3, false)
		steal = 1
	case 22 <= n && n <= 24:
		b.writeBits(uint32(265+(n/2-11)>>1)-256, 7, true)
		b.writeBits(uint32(n/2-11)&1, 1, false)
		b.writeBits(8, 5, true)
		b.writeBits(uint32(n-16-1), 3, false)
		b.writeBits(uint32(265+(n-n/2-11)>>1)-256, 7, true)
		b.writeBits(uint32(n-n/2-11)&1, 1, false)
		b.writeBits(8, 5, true)
		b.writeBits(uint32(n-16-1), 3, false)
		steal = 2
	case 25 <= n && n <= 32:
		b.writeBits(uint32(265+(n/2-11)>>1)-256, 7, true)
		b.writeBits(uint32(n/2-11)&1, 1, false)
		b.writeBits(9, 5, true)
		b.writeBits(uint32(n-24-1), 3, false)
		b.writeBits(uint32(265+(n-n/2-11)>>1)-256, 7, true)
		b.writeBits(uint32(n-n/2-11)&1, 1, false)
		b.writeBits(9, 5, true)
		b.writeBits(uint32(n-24-1), 3, false)
		steal = 2
	case 33 <= n && n <= 36:
		b.writeBits(uint32(265+(n/2-11)>>1)-256, 7, true)
		b.writeBits(uint32(n/2-11)&1, 1, false)
		b.writeBits(10, 5, true)
		b.writeBits(uint32(n-32-1), 4, false)
		b.writeBits(uint32(265+(n-n/2-11)>>1)-256, 7, true)
		b.writeBits(uint32(n-n/2-11)&1, 1, false)
		b.writeBits(10, 5, true)
		b.writeBits(uint32(n-32-1), 4, false)
		steal = 4
	case 37 <= n && n <= 48:
		b.writeBits(uint32(265+(18-11)>>1)-256, 7, true)
		b.writeBits(uint32(18-11)&1, 1, false)
		b.writeBits(10, 5, true)
		b.writeBits(uint32(n-32-1), 4, false)
		b.writeBits(uint32(269+(n-18-19)>>2)-256, 7, true)
		b.writeBits(uint32(n-18-19)&3, 2, false)
		b.writeBits(10, 5, true)
		b.writeBits(uint32(n-32-1), 4, false)
		steal = 5
	case 49 <= n && n <= 64:
		b.writeBits(uint32(254+10)-256, 7, true)
		b.writeBits(11, 5, true)
		b.writeBits(uint32(n-48-1), 4, false)
		b.writeBits(uint32(273+(n-10-35)>>3)-256, 7, true)
		b.writeBits(uint32(n-10-35)&7, 3, false)
		b.writeBits(11, 5, true)
		b.writeBits(uint32(n-48-1), 4, false)
		steal = 5
	default:
		panic("cannot encode REP")
	}
	b.writeBits(0, 7-steal, true)
}
var inflateO, inflateB int
func deflate(data []byte, litNext bool, final bool) []byte {
	var buf bytes.Buffer
	w, _ := flate.NewWriter(&buf, 9)
	w.Write(data)
	w.Close()
	z := buf.Bytes()
	if final {
		return z
	}
	b1 := bytes.NewBuffer(z)
	var b2 bytes.Buffer
	r := NewInflater(b1)
	io.Copy(&b2, r)
	r.Close()
	if inflateB == 0 {
		return z[0:inflateO]
	}
	z[inflateO] ^= 1<<uint(inflateB)
	if litNext {
		if inflateB >= 6 && len(z) == inflateO+1+5 && z[inflateO+1] == 0 && z[inflateO+2] == 0 && z[inflateO+3] == 0 && z[inflateO+4] == 0xff && z[inflateO+5] == 0xff {
			return z[0:inflateO+1]
		}
		if inflateB <= 5 && z[inflateO] == 0 {
			return z[0:inflateO]
		}
	}
	return z
}
func inflate(data []byte) []byte {
	r := NewInflater(bytes.NewBuffer(data))
	var b bytes.Buffer
	io.Copy(&b, r)
	r.Close()
	return b.Bytes()
}
const (
	maxCodeLen = 16
	maxHist    = 32768
	maxLit     = 286
	maxDist    = 32
	numCodes   = 19
)
type CorruptInputError int64
func (e CorruptInputError) String() string {
	return "flate: corrupt input before offset " + strconv.FormatInt(int64(e), 10)
}
type InternalError string
func (e InternalError) String() string { return "flate: internal error: " + string(e) }
type ReadError struct {
	Offset int64
	Error  error
}
func (e *ReadError) String() string {
	return "flate: read error at offset " + strconv.FormatInt(e.Offset, 10)
}
type WriteError struct {
	Offset int64
	Error  error
}
func (e *WriteError) String() string {
	return "flate: write error at offset " + strconv.FormatInt(e.Offset, 10)
}
type huffmanDecoder struct {
	min, max int
	limit [maxCodeLen + 1]int
	base [maxCodeLen + 1]int
	codes []int
}
func (h *huffmanDecoder) init(bits []int) bool {
	var count [maxCodeLen + 1]int
	var min, max int
	for _, n := range bits {
		if n == 0 {
			continue
		}
		if min == 0 || n < min {
			min = n
		}
		if n > max {
			max = n
		}
		count[n]++
	}
	if max == 0 {
		return false
	}
	h.min = min
	h.max = max
	code := 0
	seq := 0
	var nextcode [maxCodeLen]int
	for i := min; i <= max; i++ {
		n := count[i]
		nextcode[i] = code
		h.base[i] = code - seq
		code += n
		seq += n
		h.limit[i] = code - 1
		code <<= 1
	}
	if len(h.codes) < len(bits) {
		h.codes = make([]int, len(bits))
	}
	for i, n := range bits {
		if n == 0 {
			continue
		}
		code := nextcode[n]
		nextcode[n]++
		seq := code - h.base[n]
		h.codes[seq] = i
	}
	return true
}
var fixedHuffmanDecoder = huffmanDecoder{
	7, 9,
	[maxCodeLen + 1]int{7: 23, 199, 511},
	[maxCodeLen + 1]int{7: 0, 24, 224},
	[]int{
		256, 257, 258, 259, 260, 261, 262,
		263, 264, 265, 266, 267, 268, 269,
		270, 271, 272, 273, 274, 275, 276,
		277, 278, 279,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
		12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
		22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
		32, 33, 34, 35, 36, 37, 38, 39, 40, 41,
		42, 43, 44, 45, 46, 47, 48, 49, 50, 51,
		52, 53, 54, 55, 56, 57, 58, 59, 60, 61,
		62, 63, 64, 65, 66, 67, 68, 69, 70, 71,
		72, 73, 74, 75, 76, 77, 78, 79, 80, 81,
		82, 83, 84, 85, 86, 87, 88, 89, 90, 91,
		92, 93, 94, 95, 96, 97, 98, 99, 100,
		101, 102, 103, 104, 105, 106, 107, 108,
		109, 110, 111, 112, 113, 114, 115, 116,
		117, 118, 119, 120, 121, 122, 123, 124,
		125, 126, 127, 128, 129, 130, 131, 132,
		133, 134, 135, 136, 137, 138, 139, 140,
		141, 142, 143,
		280, 281, 282, 283, 284, 285, 286, 287,
		144, 145, 146, 147, 148, 149, 150, 151,
		152, 153, 154, 155, 156, 157, 158, 159,
		160, 161, 162, 163, 164, 165, 166, 167,
		168, 169, 170, 171, 172, 173, 174, 175,
		176, 177, 178, 179, 180, 181, 182, 183,
		184, 185, 186, 187, 188, 189, 190, 191,
		192, 193, 194, 195, 196, 197, 198, 199,
		200, 201, 202, 203, 204, 205, 206, 207,
		208, 209, 210, 211, 212, 213, 214, 215,
		216, 217, 218, 219, 220, 221, 222, 223,
		224, 225, 226, 227, 228, 229, 230, 231,
		232, 233, 234, 235, 236, 237, 238, 239,
		240, 241, 242, 243, 244, 245, 246, 247,
		248, 249, 250, 251, 252, 253, 254, 255,
	},
}
type Reader interface {
	io.Reader
	ReadByte() (c byte, err error)
}
type inflater struct {
	r       Reader
	w       io.Writer
	roffset int64
	woffset int64
	b  uint32
	nb uint
	h1, h2 huffmanDecoder
	bits     [maxLit + maxDist]int
	codebits [numCodes]int
	hist  [maxHist]byte
	hp    int
	hfull bool
	buf [4]byte
}
func (f *inflater) inflate() (err error) {
	final := false
	for err == nil && !final {
		for f.nb < 1+2 {
			if err = f.moreBits(); err != nil {
				return
			}
		}
		final = f.b&1 == 1
		f.b >>= 1
		typ := f.b & 3
		f.b >>= 2
		f.nb -= 1 + 2
		if final {
			o := int(f.roffset) - 1
			b := 8 - int(f.nb) - 3
			if b < 0 {
				o--
				b += 8
			}
			inflateO = o
			inflateB = b
		}
		switch typ {
		case 0:
			err = f.dataBlock()
		case 1:
			err = f.decodeBlock(&fixedHuffmanDecoder, nil)
		case 2:
			if err = f.readHuffman(); err == nil {
				err = f.decodeBlock(&f.h1, &f.h2)
			}
		}
	}
	return
}
var codeOrder = [...]int{16, 17, 18, 0, 8, 7, 9, 6, 10, 5, 11, 4, 12, 3, 13, 2, 14, 1, 15}
func (f *inflater) readHuffman() error {
	for f.nb < 5+5+4 {
		if err := f.moreBits(); err != nil {
			return err
		}
	}
	nlit := int(f.b&0x1F) + 257
	f.b >>= 5
	ndist := int(f.b&0x1F) + 1
	f.b >>= 5
	nclen := int(f.b&0xF) + 4
	f.b >>= 4
	f.nb -= 5 + 5 + 4
	for i := 0; i < nclen; i++ {
		for f.nb < 3 {
			if err := f.moreBits(); err != nil {
				return err
			}
		}
		f.codebits[codeOrder[i]] = int(f.b & 0x7)
		f.b >>= 3
		f.nb -= 3
	}
	for i := nclen; i < len(codeOrder); i++ {
		f.codebits[codeOrder[i]] = 0
	}
	for i, n := 0, nlit+ndist; i < n; {
		x, err := f.huffSym(&f.h1)
		if err != nil {
			return err
		}
		if x < 16 {
			f.bits[i] = x
			i++
			continue
		}
		var rep int
		var nb uint
		var b int
		switch x {
		case 16:
			rep = 3
			nb = 2
			b = f.bits[i-1]
		case 17:
			rep = 3
			nb = 3
			b = 0
		case 18:
			rep = 11
			nb = 7
			b = 0
		}
		for f.nb < nb {
			if err := f.moreBits(); err != nil {
				return err
			}
		}
		rep += int(f.b & uint32(1<<nb-1))
		f.b >>= nb
		f.nb -= nb
		for j := 0; j < rep; j++ {
			f.bits[i] = b
			i++
		}
	}
	return nil
}
func (f *inflater) decodeBlock(hl, hd *huffmanDecoder) error {
	for {
		v, err := f.huffSym(hl)
		if err != nil {
			return err
		}
		var n uint
		var length int
		switch {
		case v < 256:
			f.hist[f.hp] = byte(v)
			f.hp++
			if f.hp == len(f.hist) {
				if err = f.flush(); err != nil {
					return err
				}
			}
			continue
		case v == 256:
			return nil
		case v < 265:
			length = v - (257 - 3)
			n = 0
		case v < 269:
			length = v*2 - (265*2 - 11)
			n = 1
		case v < 273:
			length = v*4 - (269*4 - 19)
			n = 2
		case v < 277:
			length = v*8 - (273*8 - 35)
			n = 3
		case v < 281:
			length = v*16 - (277*16 - 67)
			n = 4
		case v < 285:
			length = v*32 - (281*32 - 131)
			n = 5
		default:
			length = 258
			n = 0
		}
		if n > 0 {
			for f.nb < n {
				if err = f.moreBits(); err != nil {
					return err
				}
			}
			length += int(f.b & uint32(1<<n-1))
			f.b >>= n
			f.nb -= n
		}
		var dist int
		if hd == nil {
			for f.nb < 5 {
				if err = f.moreBits(); err != nil {
					return err
				}
			}
			dist = int(reverseByte[(f.b&0x1F)<<3])
			f.b >>= 5
			f.nb -= 5
		} else {
			if dist, err = f.huffSym(hd); err != nil {
				return err
			}
		}
		switch {
		case dist < 4:
			dist++
		default:
			nb := uint(dist-2) >> 1
			extra := (dist & 1) << nb
			for f.nb < nb {
				if err = f.moreBits(); err != nil {
					return err
				}
			}
			extra |= int(f.b & uint32(1<<nb-1))
			f.b >>= nb
			f.nb -= nb
			dist = 1<<(nb+1) + 1 + extra
		}
		p := f.hp - dist
		if p < 0 {
			p += len(f.hist)
		}
		for i := 0; i < length; i++ {
			f.hist[f.hp] = f.hist[p]
			f.hp++
			p++
			if f.hp == len(f.hist) {
				if err = f.flush(); err != nil {
					return err
				}
			}
			if p == len(f.hist) {
				p = 0
			}
		}
	}
	panic("unreached")
}
func (f *inflater) dataBlock() error {
	f.nb = 0
	f.b = 0
	nr, _ := io.ReadFull(f.r, f.buf[0:4])
	f.roffset += int64(nr)
	n := int(f.buf[0]) | int(f.buf[1])<<8
	for n > 0 {
		m := len(f.hist) - f.hp
		if m > n {
			m = n
		}
		m, _ = io.ReadFull(f.r, f.hist[f.hp:f.hp+m])
		f.roffset += int64(m)
		n -= m
		f.hp += m
	}
	return nil
}
func (f *inflater) moreBits() error {
	c, err := f.r.ReadByte()
	if err != nil {
		return err
	}
	f.roffset++
	f.b |= uint32(c) << f.nb
	f.nb += 8
	return nil
}
func (f *inflater) huffSym(h *huffmanDecoder) (int, error) {
	for n := uint(h.min); n <= uint(h.max); n++ {
		lim := h.limit[n]
		if lim == -1 {
			continue
		}
		for f.nb < n {
			if err := f.moreBits(); err != nil {
				return 0, err
			}
		}
		v := int(f.b & uint32(1<<n-1))
		v <<= 16 - n
		v = int(reverseByte[v>>8]) | int(reverseByte[v&0xFF])<<8
		if v <= lim {
			f.b >>= n
			f.nb -= n
			return h.codes[v-h.base[n]], nil
		}
	}
	return 0, errors.New(strconv.FormatInt(f.roffset, 10))
}
func (f *inflater) flush() error {
	if f.hp == 0 {
		return nil
	}
	f.w.Write(f.hist[0:f.hp])
	f.woffset += int64(f.hp)
	f.hp = 0
	f.hfull = true
	return nil
}
func makeReader(r io.Reader) Reader {
	if rr, ok := r.(Reader); ok {
		return rr
	}
	return bufio.NewReader(r)
}
func (f *inflater) inflater(r io.Reader, w io.Writer) error {
	f.r = makeReader(r)
	f.w = w
	f.woffset = 0
	if err := f.inflate(); err != nil {
		return err
	}
	if err := f.flush(); err != nil {
		return err
	}
	return nil
}
func NewInflater(r io.Reader) io.ReadCloser {
	var f inflater
	pr, pw := io.Pipe()
	go func() { pw.CloseWithError(f.inflater(r, pw)) }()
	return pr
}
var reverseByte = [256]byte{
	0x00, 0x80, 0x40, 0xc0, 0x20, 0xa0, 0x60, 0xe0,
	0x10, 0x90, 0x50, 0xd0, 0x30, 0xb0, 0x70, 0xf0,
	0x08, 0x88, 0x48, 0xc8, 0x28, 0xa8, 0x68, 0xe8,
	0x18, 0x98, 0x58, 0xd8, 0x38, 0xb8, 0x78, 0xf8,
	0x04, 0x84, 0x44, 0xc4, 0x24, 0xa4, 0x64, 0xe4,
	0x14, 0x94, 0x54, 0xd4, 0x34, 0xb4, 0x74, 0xf4,
	0x0c, 0x8c, 0x4c, 0xcc, 0x2c, 0xac, 0x6c, 0xec,
	0x1c, 0x9c, 0x5c, 0xdc, 0x3c, 0xbc, 0x7c, 0xfc,
	0x02, 0x82, 0x42, 0xc2, 0x22, 0xa2, 0x62, 0xe2,
	0x12, 0x92, 0x52, 0xd2, 0x32, 0xb2, 0x72, 0xf2,
	0x0a, 0x8a, 0x4a, 0xca, 0x2a, 0xaa, 0x6a, 0xea,
	0x1a, 0x9a, 0x5a, 0xda, 0x3a, 0xba, 0x7a, 0xfa,
	0x06, 0x86, 0x46, 0xc6, 0x26, 0xa6, 0x66, 0xe6,
	0x16, 0x96, 0x56, 0xd6, 0x36, 0xb6, 0x76, 0xf6,
	0x0e, 0x8e, 0x4e, 0xce, 0x2e, 0xae, 0x6e, 0xee,
	0x1e, 0x9e, 0x5e, 0xde, 0x3e, 0xbe, 0x7e, 0xfe,
	0x01, 0x81, 0x41, 0xc1, 0x21, 0xa1, 0x61, 0xe1,
	0x11, 0x91, 0x51, 0xd1, 0x31, 0xb1, 0x71, 0xf1,
	0x09, 0x89, 0x49, 0xc9, 0x29, 0xa9, 0x69, 0xe9,
	0x19, 0x99, 0x59, 0xd9, 0x39, 0xb9, 0x79, 0xf9,
	0x05, 0x85, 0x45, 0xc5, 0x25, 0xa5, 0x65, 0xe5,
	0x15, 0x95, 0x55, 0xd5, 0x35, 0xb5, 0x75, 0xf5,
	0x0d, 0x8d, 0x4d, 0xcd, 0x2d, 0xad, 0x6d, 0xed,
	0x1d, 0x9d, 0x5d, 0xdd, 0x3d, 0xbd, 0x7d, 0xfd,
	0x03, 0x83, 0x43, 0xc3, 0x23, 0xa3, 0x63, 0xe3,
	0x13, 0x93, 0x53, 0xd3, 0x33, 0xb3, 0x73, 0xf3,
	0x0b, 0x8b, 0x4b, 0xcb, 0x2b, 0xab, 0x6b, 0xeb,
	0x1b, 0x9b, 0x5b, 0xdb, 0x3b, 0xbb, 0x7b, 0xfb,
	0x07, 0x87, 0x47, 0xc7, 0x27, 0xa7, 0x67, 0xe7,
	0x17, 0x97, 0x57, 0xd7, 0x37, 0xb7, 0x77, 0xf7,
	0x0f, 0x8f, 0x4f, 0xcf, 0x2f, 0xaf, 0x6f, 0xef,
	0x1f, 0x9f, 0x5f, 0xdf, 0x3f, 0xbf, 0x7f, 0xff,
}
func reverseUint16(v uint16) uint16 {
	return uint16(reverseByte[v>>8]) | uint16(reverseByte[v&0xFF])<<8
}
func reverseBits(number uint16, bitLength byte) uint16 {
	return reverseUint16(number << uint8(16-bitLength))
}
