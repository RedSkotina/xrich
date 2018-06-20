package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/RedSkotina/xrich"
)

func parseAndJoinJSONL(readers []io.Reader) []xrich.Record {
	var res []xrich.Record

	for _, r := range readers {
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			var rec xrich.Record

			lr := strings.NewReader(sc.Text())
			dec := json.NewDecoder(lr)
			err := dec.Decode(&rec)
			if err != nil {
				log.Println(err)
				continue
			}
			res = append(res, rec)
		}
		if err := sc.Err(); err != nil {
			log.Println("reading input:", err)
			continue
		}

	}

	return res
}

func main() {

	maxgen := flag.Int("l", xrich.MAXGEN, "number of generated words")
	question := flag.String("q", "", "Find answer for question")
	gendump := flag.Bool("d", false, "Dump state table")

	flag.Parse()
	flags := flag.Args()

	var readers []io.Reader

	for _, fpath := range flags {
		file, err := os.Open(fpath)
		if err != nil {
			log.Println(err)
			continue
		}
		reader := bufio.NewReader(file)
		readers = append(readers, reader)
	}

	recs := parseAndJoinJSONL(readers)

	c := xrich.NewChain()
	c.Build(recs)

	if *gendump {
		ioutil.WriteFile("chain.dump", []byte(c.Dump()), 0644)
	}

	if *question == "" {
		text := c.Generate(*maxgen)
		log.Println(text)
	} else {
		text := c.GenerateAnswer(*question, *maxgen)
		log.Println(text)
	}

}
