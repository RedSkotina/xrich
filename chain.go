package xrich

import (
	"bufio"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

const (
	// NPREF is Prefix length
	NPREF = 2
	// NONWORD is empty word
	NONWORD = "\n"
	// MAXGEN is max number of generated words
	MAXGEN = 50
	// SEP is separator between phrases
	SEP = "."
)

//Prefix is key for map {prefix:suffix}
type Prefix struct {
	words [NPREF]string
}

//Suffix is value for map {prefix:suffix}
type Suffix struct {
	isMarked bool
	word     string
}

func newPrefix(logger *zap.SugaredLogger, words ...string) *Prefix {
	if len(words) != NPREF {
		logger.Fatalw("tried intialize Prefex with invalid len")
	}
	prefix := Prefix{}
	for i := 0; i < NPREF; i++ {
		prefix.words[i] = words[i]
	}
	return &prefix
}

func (r *Prefix) fill(word string) {
	for i := 0; i < NPREF; i++ {
		r.words[i] = word
	}
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
	statetab map[Prefix][]Suffix
	prefix   Prefix
	policy   GeneratePolicy
	keys     []*Prefix

	logger *zap.SugaredLogger
}

//NewMarkovChain create new object of MarkovChain
func NewMarkovChain(logger *zap.Logger) MarkovChain {
	sugaredLogger := logger.Sugar()
	return MarkovChain{
		statetab: make(map[Prefix][]Suffix),
		prefix: *newPrefix(sugaredLogger, NONWORD, NONWORD),
		policy: new(RandomGeneratePolicy),
		logger: sugaredLogger,
	}
}

func (r *MarkovChain) setGeneratePolicy(p GeneratePolicy) {
	r.policy = p
}

func (r *MarkovChain) add(word string, isMarked bool) {
	suf, ok := r.statetab[r.prefix]
	if ok {
		suf = append(suf, Suffix{isMarked, word})
		r.statetab[r.prefix] = suf
	} else {
		p := Prefix{r.prefix.words}
		r.statetab[p] = []Suffix{Suffix{isMarked, word}}
		r.keys = append(r.keys, &p)
	}

	r.prefix.lshift()
	r.prefix.put(word)
}

//Build state table for markov chain with array of text blocks
func (r *MarkovChain) Build(textBlocks []string) {
	logger := r.logger.With("func", "Build")
	r.policy.init(r)
	for _, s := range textBlocks {
		rd := strings.NewReader(s)
		sc := bufio.NewScanner(rd)
		sc.Split(bufio.ScanWords)
		for sc.Scan() {
			r.add(sc.Text(), false)

			if err := sc.Err(); err != nil {
				logger.Errorw("error scanning word", err)
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
		suf = r.policy.findSuffix(sx).word
		// jump to random after phrase end
		if suf == SEP {
			r.prefix = r.policy.findNextPrefix(r)
			return suf
		}

	} else {
		return NONWORD
	}

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
	logger := r.logger.With("func", "GenerateAnswer")
	var phrases []string

	if len(r.statetab) == 0 {
		return res
	}

	prefix := *newPrefix(logger, NONWORD, SEP)

	sr := strings.NewReader(message)
	sc := bufio.NewScanner(sr)
	sc.Split(bufio.ScanWords)

	for sc.Scan() {
		w := sc.Text()
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
		logger.Errorw("error scanning", err)
		return res
	}
	if len(phrases) > 0 {
		res = r.policy.findPhrase(phrases)
	}

	return res
}
