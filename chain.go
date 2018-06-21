package xrich

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

const (
	// NPREF is Prefix length
	NPREF = 2
	// NONWORD is empty word
	NONWORD = "\n"
	// MAXGEN Number of generated words
	MAXGEN = 50
	// SEP is separator between phrases
	SEP = "."
)

//StringArray is type for suffixes
type StringArray []string

//GeneratePolicy is default policy for select prefix and suffix
type GeneratePolicy struct {
}

func newGeneratePolicy() *GeneratePolicy {
	seed := time.Now().UnixNano()
	rand.Seed(seed)
	return new(GeneratePolicy)
}

func (r *GeneratePolicy) findFirstPrefix(c *MarkovChain) Prefix {
	return *c.keys[rand.Intn(len(c.keys))]
}
func (r *GeneratePolicy) findNextPrefix(c *MarkovChain) Prefix {
	return *c.keys[rand.Intn(len(c.keys))]
}
func (r *GeneratePolicy) findSuffix(sx StringArray) string {
	return sx[rand.Intn(len(sx))]
}

//Prefix is key for map {prefix:suffix}
type Prefix struct {
	isMarked bool
	words    [NPREF]string
}

func newPrefix(nwords int) *Prefix {
	prefix := Prefix{}
	for i := 0; i < nwords; i++ {
		prefix.words[i] = NONWORD
	}
	prefix.isMarked = true
	return &prefix
}

func (r *Prefix) lshift() {
	for i := 0; i < NPREF-1; i++ {
		r.words[i] = r.words[i+1]
	}
}

func (r *Prefix) put(word string) {
	r.words[NPREF-1] = word
}

//MarkovChain is keep for state transtions
type MarkovChain struct {
	statetab map[Prefix]StringArray
	prefix   Prefix
	policy   GeneratePolicy
	keys     []*Prefix
}

//NewMarkovChain create new object of MarkovChain
func NewMarkovChain() MarkovChain {
	c := MarkovChain{}
	c.statetab = make(map[Prefix]StringArray)
	c.prefix = *newPrefix(NPREF)
	c.policy = *newGeneratePolicy()
	return c
}

func (r *MarkovChain) add(word string, isMarked bool) {

	suf, ok := r.statetab[r.prefix]
	if ok {
		suf = append(suf, word)
		r.statetab[r.prefix] = suf
	} else {
		p := Prefix{false, r.prefix.words}
		r.statetab[p] = []string{word}
		r.keys = append(r.keys, &p)
	}

	r.prefix.isMarked = isMarked
	r.prefix.lshift()
	r.prefix.put(word)

}

//Build state table for markov chain with array of text blocks
func (r *MarkovChain) Build(textBlocks []string) {
	for _, s := range textBlocks {
		rd := strings.NewReader(s)
		sc := bufio.NewScanner(rd)
		sc.Split(bufio.ScanWords)
		for sc.Scan() {
			r.add(sc.Text(), false)

			if err := sc.Err(); err != nil {
				log.Println("scan word error:", err)
			}
		}
		r.add(SEP, true)
	}

}

//Dump internal variables of  Markov chain to text
func (r *MarkovChain) Dump() string {
	return fmt.Sprintf("prefix: %v\nstatetab %v\nkeys: %v\n", r.prefix, r.statetab, r.keys)
}

func (r *MarkovChain) iterateGen() string {
	sx, ok := r.statetab[r.prefix]
	var suf string
	if ok {
		suf = r.policy.findSuffix(sx)
		// jump to random after phrase end
		if suf == SEP {
			r.prefix = r.policy.findNextPrefix(r)
			return suf
		}

	} else {
		return NONWORD
	}

	r.prefix.isMarked = false
	r.prefix.lshift()
	r.prefix.put(suf)
	return suf
}

//GenerateSentence return generated text as `string` with max number words `nwords`
func (r *MarkovChain) GenerateSentence(nwords int) (res string) {
	var recs []string
	if len(r.statetab) == 0 {
		return res
	}
	r.prefix = r.policy.findFirstPrefix(r)

	for i := 0; i < nwords; i++ {
		s := r.iterateGen()
		recs = append(recs, s)
	}
	res = strings.Join(recs, " ")
	return res
}

//GenerateAnswer return generated answer for text `message` with `nwords` max number of words or ended with SEP
func (r *MarkovChain) GenerateAnswer(message string, nwords int) (res string) {
	var phrases []string

	if len(r.statetab) == 0 {
		return res
	}

	prefix := *newPrefix(NPREF)
	prefix.put(SEP)

	sr := strings.NewReader(message)
	sc := bufio.NewScanner(sr)
	sc.Split(bufio.ScanWords)

	for sc.Scan() {
		w := sc.Text()
		prefix.isMarked = false
		prefix.lshift()
		prefix.put(w)
		r.prefix = prefix
		var recs []string
		for i, s := 0, r.iterateGen(); i < nwords && s != NONWORD && s != SEP; i, s = i+1, r.iterateGen() {
			recs = append(recs, s)
		}
		if len(recs) > 0 {
			recs = append(prefix.words[:], recs...)
			if recs[0] == SEP {
				recs = recs[1:] //Remove Sep from start of phrase . UGLY!!!
			}
			phrases = append(phrases, strings.Join(recs, " "))
		}
	}
	if err := sc.Err(); err != nil {
		log.Println("scan error:", err)
		return res
	}
	if len(phrases) > 0 {
		res = phrases[rand.Intn(len(phrases))]
	}

	return res
}
