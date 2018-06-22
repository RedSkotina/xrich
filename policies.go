package xrich

import "time"
import "math/rand"

//GeneratePolicy describe how choose elements in key moments of generation
type GeneratePolicy interface {
	init(c *MarkovChain)
	findFirstPrefix(c *MarkovChain) Prefix
	findNextPrefix(c *MarkovChain) Prefix
	findSuffix(sx []Suffix) Suffix
	findPhrase(ss []string) string
}

//RandomGeneratePolicy choose random element
type RandomGeneratePolicy struct {
	rnd *rand.Rand
}

func (r RandomGeneratePolicy) init(c *MarkovChain) {
	r.rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func (r RandomGeneratePolicy) findFirstPrefix(c *MarkovChain) Prefix {
	return *c.keys[r.rnd.Intn(len(c.keys))]
}
func (r RandomGeneratePolicy) findNextPrefix(c *MarkovChain) Prefix {
	return *c.keys[r.rnd.Intn(len(c.keys))]
}
func (r RandomGeneratePolicy) findSuffix(sx []Suffix) Suffix {
	return sx[r.rnd.Intn(len(sx))]
}

func (r RandomGeneratePolicy) findPhrase(ss []string) string {
	return ss[r.rnd.Intn(len(ss))]
}

//GetFirstElementGeneratePolicy choose always first element
type GetFirstElementGeneratePolicy struct {
	rnd *rand.Rand
}

func (r GetFirstElementGeneratePolicy) init(c *MarkovChain) {
}

func (r GetFirstElementGeneratePolicy) findFirstPrefix(c *MarkovChain) Prefix {
	return *c.keys[0]
}
func (r GetFirstElementGeneratePolicy) findNextPrefix(c *MarkovChain) Prefix {
	return *c.keys[0]
}
func (r GetFirstElementGeneratePolicy) findSuffix(sx []Suffix) Suffix {
	return sx[0]
}
func (r GetFirstElementGeneratePolicy) findPhrase(ss []string) string {
	return ss[0]
}
