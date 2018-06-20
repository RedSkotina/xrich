package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io"
	"log"
	"math/rand"
	"os"
	"reflect"
	"strings"

	"strconv"

	"github.com/RedSkotina/xrich"
	"github.com/Syfaro/telegram-bot-api"
)

var (
	// глобальная переменная в которой храним токен
	telegramBotToken  string
	maxgen            int
	answerProbability float64
)

const (
	//DefaultAnswerProbability is used when env and flag is not set
	DefaultAnswerProbability = 0.25
)

func init() {
	nwords, err := strconv.Atoi(os.Getenv("XRICH_MAX_WORDS"))
	if err != nil {
		nwords = xrich.MAXGEN
	}

	prob, err := strconv.ParseFloat(os.Getenv("XRICH_ANSWER_PROBABALITY"), 64)
	if err != nil {
		prob = DefaultAnswerProbability
	}
	flag.StringVar(&telegramBotToken, "token", os.Getenv("XRICH_TELEGRAM_TOKEN"), "Telegram Bot Token")
	flag.IntVar(&maxgen, "max", nwords, "max number of generated words")
	flag.Float64Var(&answerProbability, "p", prob, "answer probability")
	flag.Parse()

	// без него не запускаемся
	if telegramBotToken == "" {
		log.Print("token is required")
		os.Exit(1)
	}
}

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
	filenamesEnv := os.Getenv("XRICH_INPUT_FILES")
	filenames := strings.Split(filenamesEnv, ";")
	flags := flag.Args()

	filenames = append(filenames, flags...)

	var readers []io.Reader

	for _, fpath := range filenames {
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

	// используя токен создаем новый инстанс бота
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// u - структура с конфигом для получения апдейтов
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// используя конфиг u создаем канал в который будут прилетать новые сообщения
	updates, err := bot.GetUpdatesChan(u)

	// discard all pending messages
	//updates.Clear()

	// в канал updates прилетают структуры типа Update
	// вычитываем их и обрабатываем
	for update := range updates {
		if update.Message == nil {
			continue
		}

		// логируем от кого какое сообщение пришло
		//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {
			if rand.Float64() <= answerProbability {
				reply := c.GenerateAnswer(update.Message.Text, maxgen)
				// создаем ответное сообщение
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
				// отправляем
				bot.Send(msg)
			}
		}
	}
}
