package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/RedSkotina/xrich"
)

//Record is structure represent text block from JSON
type Record struct {
	Date int64  `json:"date"`
	Text string `json:"text"`
}

func parseJSONL(r io.Reader) (res []string) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		var rec Record

		lr := strings.NewReader(sc.Text())
		dec := json.NewDecoder(lr)
		err := dec.Decode(&rec)

		if err != nil {
			log.Println("parse jsonl", err)
			continue
		}

		if err := sc.Err(); err != nil {
			log.Println("scan jsonl:", err)
			continue
		}

		res = append(res, rec.Text)

	}
	return res
}

func joinInputs(readers []io.Reader) (res []string) {
	for _, r := range readers {
		ss := parseJSONL(r)
		res = append(res, ss...)
	}
	return res
}

func newReaders(filepathes []string) []io.Reader {
	var readers []io.Reader

	for _, fpath := range filepathes {
		file, err := os.Open(fpath)
		if err != nil {
			log.Println(err)
			continue
		}
		r := bufio.NewReader(file)
		readers = append(readers, r)
	}

	return readers
}

func main() {

	maxgen := flag.Int("l", xrich.MAXGEN, "number of generated words")
	question := flag.String("q", "", "Find answer for question")
	gendump := flag.Bool("d", false, "Dump state table")

	flag.Parse()

	flags := flag.Args()

	rs := newReaders(flags)
	t := joinInputs(rs)

	c := xrich.NewMarkovChain()
	c.Build(t)

	if *gendump {
		ioutil.WriteFile("markovchain.dump", []byte(c.Dump()), 0644)
	}

	if *question == "" {
		text := c.GenerateSentence(*maxgen)
		fmt.Println(text)
	} else {
		text := c.GenerateAnswer(*question, *maxgen)
		fmt.Println(text)
	}

}
