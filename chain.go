package xrich

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

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

var reClearTrash = regexp.MustCompile(`[^\p{L}\p{P}\d\s]`)
var reIsWord = regexp.MustCompile(`[\p{L}\d]`)
var reMultiPunct = regexp.MustCompile(`(\p{P}\s){2,}`)

func clearString(s string) string {
	return reClearTrash.ReplaceAllString(s, "")
}

// ScanWordsAndPunct is a split function for a Scanner that returns each
// space or punctuation separated word or punctuation  , with surrounding spaces deleted. It will
// never return an empty string. The definition of space is set by
// unicode.IsSpace. The definition of punct is set by unicode.IsPunct.
func ScanWordsAndPunct(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip leading spaces.
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !unicode.IsSpace(r) {
			break
		}
	}
	// Scan until space, marking end of word.
	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		if unicode.IsPunct(r) && i != start {
			return i, data[start:i], nil
		}
		if unicode.IsSpace(r) {
			return i + width, data[start:i], nil
		}
	}
	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}
	// Request more data.
	return start, nil, nil
}

// ScanOnlyWords is a split function for a Scanner that returns each
// space or punctuation separated word , with surrounding spaces and punctuation deleted. It will
// never return an empty string. The definition of space is set by
// unicode.IsSpace. The definition of punct is set by unicode.IsPunct.
func ScanOnlyWords(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip leading spaces.
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !unicode.IsSpace(r) && !unicode.IsPunct(r) {
			break
		}
	}
	// Scan until space, marking end of word.
	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			return i + width, data[start:i], nil
		}
	}
	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}
	// Request more data.
	return start, nil, nil
}

//Prefix is key for map {prefix:suffix}
type Prefix struct {
	words [NPREF]string
}

//Suffix is value for map {prefix:suffix}
type Suffix struct {
	sol  bool //start-of-line
	word string
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

//Context keep current state
type Context struct {
	prefix      Prefix
	preLastWord string
}

//MarkovChain are main structure that hold states transitions
type MarkovChain struct {
	statetab map[Prefix][]Suffix
	policy   GeneratePolicy
	keys     []*Prefix
	logger   *zap.SugaredLogger
}

//NewMarkovChain create new object of MarkovChain
func NewMarkovChain(logger *zap.Logger) MarkovChain {
	sugaredLogger := logger.Sugar()
	return MarkovChain{
		statetab: make(map[Prefix][]Suffix),
		policy:   new(RandomGeneratePolicy),
		logger:   sugaredLogger,
	}
}

//SetGeneratePolicy allow change choice policy of elements in key transitions
func (r *MarkovChain) SetGeneratePolicy(p GeneratePolicy) {
	r.policy = p
}

func isWord(s string) bool {
	return reIsWord.FindStringIndex(s) != nil
}

func (r *MarkovChain) stepBuild(ctx *Context, word string, sol bool) {

	r.addWord(ctx, word, sol)

	// if "a , [, b] c" then we add [a b] with same suffix c
	if ctx.preLastWord != "" && !isWord(ctx.prefix.words[0]) && isWord(ctx.prefix.words[NPREF-1]) {
		ctx.prefix.words[0] = ctx.preLastWord
		r.addWord(ctx, word, sol)
		ctx.preLastWord = ""
	}

	// example: "[? a] ."
	if isWord(ctx.prefix.words[NPREF-1]) && !isWord(word) {
		ctx.preLastWord = ctx.prefix.words[NPREF-1]
	}

	ctx.prefix.lshift()
	ctx.prefix.put(word)
}

//Add state in states transitions table and mark/unmark him as start of line using `sol`
func (r *MarkovChain) addWord(ctx *Context, word string, sol bool) {

	suf, ok := r.statetab[ctx.prefix]
	if ok {
		suf = append(suf, Suffix{sol, word})
		r.statetab[ctx.prefix] = suf
	} else {
		p := Prefix{ctx.prefix.words}
		r.statetab[p] = []Suffix{Suffix{sol, word}}
		r.keys = append(r.keys, &p)
	}

}

//Build states transition table for markov chain from text blocks
func (r *MarkovChain) Build(textBlocks []string) {
	logger := r.logger.With("func", "Build")
	ctx := new(Context)
	ctx.prefix = *newPrefix(logger, NONWORD, NONWORD)
	r.policy.init(r)
	// TODO: split punctuation?

	for i, s := range textBlocks {
		s = clearString(s)
		rd := strings.NewReader(s)
		sc := bufio.NewScanner(rd)
		sc.Split(ScanWordsAndPunct)
		for sc.Scan() {
			sol := true
			if i >= NPREF {
				sol = false
			}
			r.stepBuild(ctx, sc.Text(), sol)

		}
		if err := sc.Err(); err != nil {
			logger.Errorw("error scanning word", err)
		}
		r.stepBuild(ctx, NONWORD, false)
	}
}

//Dump internal variables of  Markov chain to text
func (r *MarkovChain) Dump() string {
	return fmt.Sprintf("statetab %v\nkeys: %v\n", r.statetab, r.keys)
}

//generationStep generate one word for context `ctx` and update context
func (r *MarkovChain) generationStep(ctx *Context) string {
	sx, ok := r.statetab[ctx.prefix]
	if !ok {
		return NONWORD
	}

	suf := r.policy.findSuffix(sx).word

	if suf != NONWORD {
		ctx.prefix.lshift()
		ctx.prefix.put(suf)
	} else {
		// phrase is ended
		ctx.prefix = r.policy.findNextPrefix(r)
		suf = SEP
	}

	return suf
}

//GenerateSentence return generated text as `string` with max number of words `nwords`
func (r *MarkovChain) GenerateSentence(nwords int) (res string) {
	if len(r.statetab) == 0 {
		return res
	}
	r.policy.init(r)

	var words []string
	ctx := new(Context)
	ctx.prefix = r.policy.findFirstPrefix(r)

	for i := 0; i < nwords; i++ {
		s := r.generationStep(ctx)
		words = append(words, s)
	}

	res = strings.Join(words, " ")
	res = reMultiPunct.ReplaceAllString(res, "$1")
	return res
}

//GenerateAnswer return generated answer for text `message` with max number of words `nwords` or ended with NONWORD/SEP
func (r *MarkovChain) GenerateAnswer(message string, nwords int) (res string) {
	logger := r.logger.With("func", "GenerateAnswer")

	if len(r.statetab) == 0 {
		return res
	}
	r.policy.init(r)

	var phrases []string

	prefix := *newPrefix(logger, NONWORD, NONWORD)

	sr := strings.NewReader(message)
	sc := bufio.NewScanner(sr)
	sc.Split(ScanOnlyWords)
	for sc.Scan() {
		w := sc.Text()

		prefix.lshift()
		prefix.put(w)

		ctx := new(Context)
		ctx.prefix = prefix

		var words []string
		for i, s := 0, r.generationStep(ctx); i < nwords && s != NONWORD && s != SEP; i, s = i+1, r.generationStep(ctx) {
			words = append(words, s)
		}
		if len(words) > 0 {
			//remove nonword from start
			k := 0
			for i := 0; i < NPREF && prefix.words[i] == NONWORD; i++ {
				k++
			}
			words = append(prefix.words[k:NPREF], words...)
			s := strings.Join(words, " ")
			s = reMultiPunct.ReplaceAllString(s, "$1")
			phrases = append(phrases, s)
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
