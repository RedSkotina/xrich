package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/RedSkotina/xrich"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/telegram-bot-api.v4"
)

var (
	logger *zap.SugaredLogger
)

func init() {
	// FLAG (PRIMARY):
	flag.String("token", "", "Telegram Bot Token")
	flag.Int("maxwords", xrich.MAXGEN, "number of generated words")
	flag.Int("answerProbability", xrich.MAXGEN, "answer probabality")
	flag.Bool("logjson", false, "log to json")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	// ENV (SECONDARY):
	//viper.SetEnvPrefix("XRICH") //viper bug: cant use symbol _ in BindEnv
	//viper.AutomaticEnv()
	viper.BindEnv("token", "XRICH_TELEGRAM_TOKEN")
	viper.BindEnv("maxwords", "XRICH_MAX_WORDS")
	viper.BindEnv("answerProbability", "XRICH_ANSWER_PROBABALITY")
	viper.BindEnv("infiles", "XRICH_INPUT_FILES")

	// DEFAULT:
	viper.SetDefault("token", "")
	viper.SetDefault("maxwords", xrich.MAXGEN)
	viper.SetDefault("answerProbability", 0.25)

	// PARSE:
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	loggerCfg := zap.NewProductionConfig()
	if viper.GetBool("logtojson") {
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

	if viper.GetString("token") == "" {
		logger.Fatalw("token is required")
	}
}

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
			logger.Errorw("failed to decode jsonl", err)
			continue
		}

		if err := sc.Err(); err != nil {
			logger.Errorw("failed to scan jsonl", err)
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
			logger.Errorw("failed to open file",
				"path", fpath,
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
	var filenames []string

	filenamesEnv := viper.GetString("infiles")
	if filenamesEnv != "" {
		filenames = strings.Split(filenamesEnv, ";")
	}
	flags := pflag.Args()
	filenames = append(filenames, flags...)

	rs := newReaders(filenames)
	t := joinInputs(rs)

	c := xrich.NewMarkovChain(logger.Desugar())
	c.Build(t)

	bot, err := tgbotapi.NewBotAPI(viper.GetString("token"))
	if err != nil {
		logger.Fatalw("failed to initialize botapi", err)
	}

	logger.Infow("authorized on account",
		"account", bot.Self.UserName,
	)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	// discard all pending messages
	// Optional: wait for updates and clear them if you don't want to handle
	// a large backlog of old messages
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	for update := range updates {
		if update.Message == nil {
			continue
		}

		//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.Text != "" {
			if rand.Float64() <= viper.GetFloat64("answerProbability") {
				reply := c.GenerateAnswer(update.Message.Text, viper.GetInt("maxwords"))
				if reply != "" {
					_, err = bot.Send(tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping))
					if err != nil {
						logger.Warnw("unable to send 'typing' status to the channel", err)
					}
					waitTime := time.Duration((rand.Int()%100000*len(reply))%2000) * time.Millisecond
					time.Sleep(waitTime)

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)

					_, err = bot.Send(msg)
					if err != nil {
						logger.Errorw("unable to send error message", err)
					}
				}
			}
		}
	}
}
