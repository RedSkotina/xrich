package xrich

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	logger, _ = zap.NewDevelopment()
}

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
	c := NewMarkovChain(logger)
	c.SetGeneratePolicy(testGeneratePolicy{})
	c.Build(ss)
	s := c.GenerateSentence(3)
	assert.Equal(t, "a b c", s)
}

func TestGenerate2(t *testing.T) {
	ss := []string{"a b c b", "b c d"}
	c := NewMarkovChain(logger)
	c.SetGeneratePolicy(testGeneratePolicy{})
	c.Build(ss)
	s := c.GenerateSentence(6)
	assert.Equal(t, "a b c b . a", s)
}

func TestAnswer1(t *testing.T) {
	ss := []string{"a b c b", "b c d"}
	c := NewMarkovChain(logger)
	c.SetGeneratePolicy(testGeneratePolicy{})
	c.Build(ss)
	s := c.GenerateAnswer("a", 6)
	assert.Equal(t, "a b c b", s)
}

func TestAnswer2(t *testing.T) {
	ss := []string{"a b c b", "b c d"}
	c := NewMarkovChain(logger)
	c.SetGeneratePolicy(testGeneratePolicy{})
	c.Build(ss)
	s := c.GenerateAnswer("b", 6)
	assert.Equal(t, "b c b", s)
}

func TestAnswer3(t *testing.T) {
	ss := []string{"\u2318a, b: c- b.", "b c d"}
	c := NewMarkovChain(logger)
	c.SetGeneratePolicy(testGeneratePolicy{})
	c.Build(ss)
	fmt.Println(c.Dump())
	s := c.GenerateAnswer("b", 6)
	assert.Equal(t, "b c - b", s)
}

func TestAnswer4(t *testing.T) {
	ss := []string{"a, .  b c b . .", "b c d"}
	c := NewMarkovChain(logger)
	c.SetGeneratePolicy(testGeneratePolicy{})
	c.Build(ss)
	s := c.GenerateAnswer("a", 10)
	assert.Equal(t, "a ,", s)
}

func TestAnswer5(t *testing.T) {
	ss := []string{"a, .  b c b . .", "b c d"}
	c := NewMarkovChain(logger)
	c.SetGeneratePolicy(testGeneratePolicy{})
	c.Build(ss)
	s := c.GenerateAnswer("b,c", 10)
	assert.Equal(t, "b c b", s)
}

func TestAnswerR1(t *testing.T) {
	ss := []string{"так он сидел в девопсе", "прод с девопс и шлюхами за 30к", "готовых за 120к в москве это обслуживать"}
	c := NewMarkovChain(logger)
	c.SetGeneratePolicy(testGeneratePolicy{})
	c.Build(ss)
	s := c.GenerateAnswer("в девопс чатике", 50)
	assert.Equal(t, "b c b", s)
}

func TestAnswerR2(t *testing.T) {
	ss := []string{"и всё за 30к"}
	c := NewMarkovChain(logger)
	c.SetGeneratePolicy(testGeneratePolicy{})
	c.Build(ss)
	s := c.GenerateAnswer("и все", 50)
	assert.Equal(t, "и всё за 30к", s)
}

func TestAnswerR3(t *testing.T) {
	ss := []string{"кодер, ему 42 года", "и всё пиздец, ему платят"}
	c := NewMarkovChain(logger)
	c.SetGeneratePolicy(testGeneratePolicy{})
	c.Build(ss)
	s := c.GenerateAnswer("и всё", 50)
	assert.Equal(t, "и всё пиздец , ему 42 года", s)
}
