package xrich

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate1(t *testing.T) {
	ss := []string{"a b c"}
	c := NewMarkovChain()
	c.setGeneratePolicy(GetFirstElementGeneratePolicy{})
	c.Build(ss)
	s := c.GenerateSentence(3)
	assert.Equal(t, "a b c", s)
}

func TestGenerate2(t *testing.T) {
	ss := []string{"a b c b", "b c d"}
	c := NewMarkovChain()
	c.setGeneratePolicy(GetFirstElementGeneratePolicy{})
	c.Build(ss)
	s := c.GenerateSentence(4)
	assert.Equal(t, "a b c", s)
}
