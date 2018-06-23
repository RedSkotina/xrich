package xrich

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

//testGeneratePolicy is mock
type testGeneratePolicy struct {
}

func (r testGeneratePolicy) init(c *MarkovChain) {
}

func (r testGeneratePolicy) findFirstPrefix(c *MarkovChain) Prefix {
	return *c.keys[0]
}
func (r testGeneratePolicy) findNextPrefix(c *MarkovChain) Prefix {
	return *c.keys[0]
}
func (r testGeneratePolicy) findSuffix(sx []Suffix) Suffix {
	return sx[0]
}
func (r testGeneratePolicy) findPhrase(ss []string) string {
	return ss[0]
}

func TestGenerate1(t *testing.T) {
	ss := []string{"a b c"}
	c := NewMarkovChain()
	c.SetGeneratePolicy(testGeneratePolicy{})
	c.Build(ss)
	s := c.GenerateSentence(3)
	assert.Equal(t, "a b c", s)
}

func TestGenerate2(t *testing.T) {
	ss := []string{"a b c b", "b c d"}
	c := NewMarkovChain()
	c.SetGeneratePolicy(testGeneratePolicy{})
	c.Build(ss)
	s := c.GenerateSentence(6)
	assert.Equal(t, "a b c b . a", s)
}

func TestAnswer1(t *testing.T) {
	ss := []string{"a b c b", "b c d"}
	c := NewMarkovChain()
	c.SetGeneratePolicy(testGeneratePolicy{})
	c.Build(ss)
	s := c.GenerateAnswer("a", 6)
	assert.Equal(t, "a b c b", s)
}

func TestAnswer2(t *testing.T) {
	ss := []string{"a b c b", "b c d"}
	c := NewMarkovChain()
	c.SetGeneratePolicy(testGeneratePolicy{})
	c.Build(ss)
	s := c.GenerateAnswer("b", 6)
	assert.Equal(t, "b c b", s)
}

func TestAnswer3(t *testing.T) {
	ss := []string{"\u2318a, b: c- b.", "b c d"}
	c := NewMarkovChain()
	c.SetGeneratePolicy(testGeneratePolicy{})
	c.Build(ss)
	s := c.GenerateAnswer("b", 6)
	assert.Equal(t, "b c b", s)
}

func TestAnswer4(t *testing.T) {
	ss := []string{"a, .  b c b . .", "b c d"}
	c := NewMarkovChain()
	c.SetGeneratePolicy(testGeneratePolicy{})
	c.Build(ss)
	fmt.Println(c.Dump())
	s := c.GenerateAnswer("a", 10)
	assert.Equal(t, "a ,", s)
}

func TestAnswer5(t *testing.T) {
	ss := []string{"a, .  b c b . .", "b c d"}
	c := NewMarkovChain()
	c.SetGeneratePolicy(testGeneratePolicy{})
	c.Build(ss)
	fmt.Println(c.Dump())
	s := c.GenerateAnswer("b,c", 10)
	assert.Equal(t, "b c b", s)
}
