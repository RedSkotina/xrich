# Xrich

Generate random text from original text using Markov chains

## How to use?

1. Import the package:

`import "github.com/RedSkotina/xrich"`

2. Create the `Chain` Struct

`c := xrich.NewChain()`

3. Fill `[]xrich.Record` with text blocks.

Look `parseAndJoinJSONL(readers []io.Reader) []xrich.Record` from `cmd/xrich/main.go` for example.

4. Pass  `[]xrich.Record` to Build for initializing internal state table of the `Chain`

`c.Build(recs)`

5. Call `Generate` method with Maximum word count `MAXGEN`    

`t := c.Generate(MAXGEN)`

or

6. Call `GenerateAnswer` method with trigger message `message`

`t := c.GenerateAnswer(message, MAXGEN)`
