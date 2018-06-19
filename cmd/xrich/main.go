package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"strings"

	"github.com/RedSkotina/xrich"
)

func parseAndJoinJSON(readers []io.Reader) []xrich.Record {
	var res []xrich.Record

	for _, rd := range readers {
		var recs []xrich.Record
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

	maxgen := flag.Int("l", xrich.MAXGEN, "number of generated words")
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

	c := xrich.newChain()
	c.build(recs)
	t := c.generate(*maxgen)
	text := strings.Join(t, " ")
	log.Println(text)

}
