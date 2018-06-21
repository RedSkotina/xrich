package xrich

import "testing"
import "hash/fnv"
import "fmt"
import "strings"

type Prefix0 struct {
	words [2]string
}

// temp variable for avoid optimizing-out result
var HasKey bool

var m0 map[Prefix0]int

func BenchmarkHashmapSearchEmbedStruct(b *testing.B) {
	p := Prefix0{[2]string{"строка1", "строка2"}}
	m0 = make(map[Prefix0]int)
	m0[p] = 1
	for i := 0; i < b.N; i++ {
		_, HasKey = m0[p]
	}
}

type PrefixH struct {
	words [2]string
}

var muint32 map[uint32]int

func (r *PrefixH) hashfn1() uint32 {
	var h uint32 = 7
	const multiplier = 31
	const nhash = 131071
	for _, w := range r.words {
		for _, e := range w {
			h += h*multiplier + uint32(e)
		}
	}
	return h % nhash
}

func getKeyUint32(b *testing.B, hashfn func() uint32) {
	muint32 = make(map[uint32]int)
	muint32[hashfn()] = 1
	for i := 0; i < b.N; i++ {
		_, HasKey = muint32[hashfn()]
	}
}

func BenchmarkHashmapSearchHashfn1(b *testing.B) {
	p := PrefixH{[2]string{"строка1", "строка2"}}
	getKeyUint32(b, p.hashfn1)
}

func (r *PrefixH) hashfn2() uint32 {
	h := fnv.New32a()
	for _, w := range r.words {
		h.Write([]byte(w))
	}
	return h.Sum32()
}

func BenchmarkHashmapSearchHashfn2(b *testing.B) {
	p := PrefixH{[2]string{"строка1", "строка2"}}
	getKeyUint32(b, p.hashfn2)
}

var mstring map[string]int

func getKeyString(b *testing.B, hashfn func() string) {
	mstring = make(map[string]int)
	mstring[hashfn()] = 1
	for i := 0; i < b.N; i++ {
		_, HasKey = mstring[hashfn()]
	}
}

func (r *PrefixH) hashfn3() string {
	h := fmt.Sprintf("%s%s", r.words[0], r.words[1])
	return h
}

func BenchmarkHashmapSearchHashfn3(b *testing.B) {
	p := PrefixH{[2]string{"строка1", "строка2"}}
	getKeyString(b, p.hashfn3)
}

var sb = new(strings.Builder)

func (r *PrefixH) hashfn4() string {
	sb.Reset()
	l := len(r.words)
	for k := 0; k < l; k++ {
		sb.WriteString(r.words[k])
	}
	return sb.String()
}

func BenchmarkHashmapSearchHashfn4(b *testing.B) {
	p := PrefixH{[2]string{"строка1", "строка2"}}
	getKeyString(b, p.hashfn4)
}

func (r *PrefixH) hashfn5() string {
	var s string
	l := len(r.words)
	for k := 0; k < l; k++ {
		s = s + r.words[k]
	}
	return s
}

func BenchmarkHashmapSearchHashfn5(b *testing.B) {
	p := PrefixH{[2]string{"строка1", "строка2"}}
	getKeyString(b, p.hashfn5)
}
