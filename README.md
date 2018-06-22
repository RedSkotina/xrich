# Xrich 

This is a Go package that uses Markov chains to generate random text.

## How to use?

1. Import the package:

`import "github.com/RedSkotina/xrich"`

2. Create the `MarkovChain` Struct

`c := xrich.NewMarkovChain()`

3. Fill variable of type `[]string` with text blocks represent logical pieces of text

`textBlocks := []string{"string1","string2"}`

4. Pass variable to function `Build` for initializing internal state table of the `MarkovChain`

`c.Build(textBlocks)`

5. Call `GenerateSentense` method with maximum word number `MAXGEN`    

`s := c.GenerateSentence(MAXGEN)`

or

6. Call `GenerateAnswer` method with trigger message `message` and maximum word number `MAXGEN`

`s := c.GenerateAnswer(message, MAXGEN)`


# Xrich-telebot

This is a Telegram bot, which reacts on all messages in chat and sends generated text.

## How to use?

`xrich_telebot -token=TELEGRAM_BOT_TOKEN -max=MAXWORDS file1.jsonl file2.jsonl ...`
