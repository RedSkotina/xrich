package xrich

import (
	"bufio"
	"encoding/json"
	"flag"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/RedSkotina/xrich"
)

const (
	// NPREF is Prefix length
	NPREF = 2
	// NONWORD is empty word
	NONWORD = "\n"
	// MAXGEN Number of generated words
	MAXGEN = 50
)

//Record contain a original text block
type Record struct {
	Date string `json:"date"`
	Text string `json:"text"`
}

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

func (r *GeneratePolicy) findFirstPrefix(statetab map[Prefix]StringArray) Prefix {
	return *newPrefix(NPREF)
}
func (r *GeneratePolicy) findNextPrefix(statetab map[Prefix]StringArray) Prefix {
	return *newPrefix(NPREF)
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

//Chain are store for state transtions
type Chain struct {
	statetab map[Prefix]StringArray
	prefix   Prefix
	policy   GeneratePolicy
}

//NewChain is create Markov chain
func NewChain() Chain {
	c := Chain{}
	c.statetab = make(map[Prefix]StringArray)
	c.prefix = *newPrefix(NPREF)
	c.policy = *newGeneratePolicy()
	return c
}

func (r *Chain) add(word string, isMarked bool) {
	suf, ok := r.statetab[r.prefix]
	if !ok {
		suf = []string{}
		p := Prefix{false, r.prefix.words}
		r.statetab[p] = suf
	}
	suf = append(suf, word)
	r.statetab[r.prefix] = suf

	r.prefix.isMarked = isMarked
	r.prefix.lshift()
	r.prefix.put(word)

}

func (r *Chain) build(recs []Record) {
	for _, rec := range recs {
		reader := strings.NewReader(rec.Text)
		scanner := bufio.NewScanner(reader)
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			r.add(scanner.Text(), false)
		}
		if err := scanner.Err(); err != nil {
			log.Panicln("reading input:", err)
		}
		r.add(NONWORD, true)
	}

}

func (r *Chain) generate(nwords int) []string {
	var res []string
	r.prefix = r.policy.findFirstPrefix(r.statetab)

	for i := 0; i < nwords; i++ {
		sx, ok := r.statetab[r.prefix]
		var suf string
		if ok {
			suf = r.policy.findSuffix(sx)
			if suf == NONWORD {
				r.prefix = r.policy.findNextPrefix(r.statetab)
				continue
			}
			//log.Println(suf)
			res = append(res, suf)

		} else {
			//log.Printf("not found prefix")
			continue
		}

		if r.prefix.isMarked {
			r.prefix.isMarked = false
		}
		r.prefix.lshift()
		r.prefix.put(suf)

	}
	return res
}

func parseAndJoinJSON(readers []io.Reader) []Record {
	var res []Record

	for _, rd := range readers {
		var recs []Record
		dec := json.NewDecoder(rd)
		err := dec.Decode(&recs)
		if err != nil {
			log.Println(err)
			continue
		}
		res = append(res, recs...)
	}

	return res
}

func main() {

	maxgen := flag.Int("l", MAXGEN, "number of generated words")
	flag.Parse()
	flags := flag.Args()

	var readers []io.Reader

	for _, fpath := range flags {

		file, err := os.Open(fpath)
		if err != nil {
			log.Println(err)
		}
		reader := bufio.NewReader(file)
		readers = append(readers, reader)
	}

	recs := parseAndJoinJSON(readers)

	c := xrich.NewChain()
	c.build(recs)
	t := c.generate(*maxgen)
	text := strings.Join(t, " ")
	log.Println(text)

}
