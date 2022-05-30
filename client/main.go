package main

import (
	. "github.com/cbytensky/bloxroute/common"

	"bufio"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

var smbi = sqs.SendMessageBatchInput{ // reused in SendMessageBatch
	QueueUrl: &QueueUrl,
	Entries:  make([]types.SendMessageBatchRequestEntry, 0, BatchSize),
}
var entries = &smbi.Entries

func main() {

	CommonInit()

	// Adding commands from command line
	for _, arg := range os.Args[1:] {
		addtobatch(arg)
	}

	// Adding commands from stdin if it is pipe
	stdin := os.Stdin
	stat, err := stdin.Stat()
	PanicIfErr(err)
	if stat.Mode()&os.ModeCharDevice == 0 { // Checking stdin is not TTY
		scanner := bufio.NewScanner(stdin)
		for {
			scanner.Scan()
			text := scanner.Text()
			if len(text) == 0 {
				break
			}
			addtobatch(text)
		}
	}

	// Sending last batch
	if len(*entries) > 0 {
		sendbatch()
	}
}

var entry *types.SendMessageBatchRequestEntry // used also in addAttribute

func addtobatch(line string) {
	index := len(*entries)
	*entries = append(*entries, types.SendMessageBatchRequestEntry{
		Id:                DigitToStr(index),                               // '0' .. '9'
		MessageAttributes: make(map[string]types.MessageAttributeValue, 2), // no more than 2 attributes
		MessageBody:       aws.String(" "),                                 // must not me empty
	})
	entry = &(*entries)[index]
	command := line[0] // +, -, ., !
	valid := true      // Input validation
	method := "GetAllItems"
	if len(line) == 1 {
		valid = valid && command == '!'
	} else {
		namevalue := line[1:]
		if command == '+' {
			method = "AddItem"
			// Dividing namevalue to name and value by colon
			pos := strings.IndexByte(namevalue, ':')
			valid = pos >= 0
			if valid {
				entry.MessageBody = aws.String(namevalue[pos+1:])
				namevalue = namevalue[:pos]
			}
		} else if command == '-' {
			method = "RemoveItem"
		} else if command == '.' {
			method = "GetItem"
		} else {
			valid = false
		}
		addAttribute("name", namevalue)
	}
	if !valid {
		LogErr("Malformed command: %s", line)
		return
	}
	addAttribute("method", method)

	// Sending if batch is full
	if index+1 == BatchSize {
		sendbatch()
	}
}

var refString = aws.String("String") // reused as DataType

func addAttribute(name string, value string) {
	entry.MessageAttributes[name] = types.MessageAttributeValue{
		DataType:    refString,
		StringValue: aws.String(value),
	}
}

func sendbatch() {
	_, err := SqsClient.SendMessageBatch(Context, &smbi)
	PanicIfErr(err)
	Log("INF: Messages sent: %d", len(*entries))
}
