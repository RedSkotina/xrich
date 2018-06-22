package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/RedSkotina/xrich"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

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
			logger.Errorw("error parsing jsonl", err)
			continue
		}

		if err := sc.Err(); err != nil {
			logger.Errorw("error scanning jsonl", err)
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
			logger.Errorw("error opening file",
				"file", fpath,
				err,
			)
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
	logToJson := flag.Bool("logjson", false, "log to json")

	flag.Parse()

	loggerCfg := zap.NewProductionConfig()
        if *logToJson {
                loggerCfg.Encoding = "json"
        } else {
                encoderConfig := zap.NewProductionEncoderConfig()
                encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
                loggerCfg.EncoderConfig = encoderConfig
                loggerCfg.Encoding = "console"
        }
        l, _ := loggerCfg.Build()
        _ = zap.RedirectStdLog(l)
        logger = l.Sugar()

	flags := flag.Args()

	rs := newReaders(flags)
	t := joinInputs(rs)
	if len(t) == 0 {
		logger.Fatalw("no valid input files specified")
	}

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
