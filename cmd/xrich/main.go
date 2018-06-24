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
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
	// FLAG (PRIMARY):
	flag.Int("maxwords", xrich.MAXGEN, "number of generated words")
	flag.String("question", "", "find answer for question")
	flag.Bool("gendump", false, "dump state table")
	flag.Bool("logjson", false, "log to json")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	// ENV (SECONDARY):
	//viper.SetEnvPrefix("XRICH")

	viper.BindEnv("maxwords", "XRICH_MAX_WORDS")

	// DEFAULT:
	viper.SetDefault("maxwords", xrich.MAXGEN)

	// PARSE:
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	flags := pflag.Args()

	loggerCfg := zap.NewProductionConfig()
	if viper.GetBool("logjson") {
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

	rs := newReaders(flags)
	t := joinInputs(rs)
	if len(t) == 0 {
		logger.Fatalw("no valid input files specified")
	}

	c := xrich.NewMarkovChain(logger.Desugar())
	c.Build(t)

	if viper.GetBool("gendump") {
		ioutil.WriteFile("markovchain.dump", []byte(c.Dump()), 0644)
	}

	if viper.GetString("question") == "" {
		text := c.GenerateSentence(viper.GetInt("maxwords"))
		fmt.Println(text)
	} else {
		text := c.GenerateAnswer(viper.GetString("question"), viper.GetInt("maxwords"))
		fmt.Println(text)
	}

}
