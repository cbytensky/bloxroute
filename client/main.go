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

var smbi = sqs.SendMessageBatchInput{
	QueueUrl: &QueueUrl,
	Entries:  make([]types.SendMessageBatchRequestEntry, 0),
}

var entryindex uint8 // index inside batch, 0..9

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
	if entryindex > 0 {
		sendbatch()
	}
}

func addAttribute(entry types.SendMessageBatchRequestEntry, name string, value string) {
	entry.MessageAttributes[name] = types.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: &value,
	}
}

func addtobatch(line string) {

	entry := types.SendMessageBatchRequestEntry{
		Id:                DigitToStr(entryindex),
		MessageAttributes: make(map[string]types.MessageAttributeValue),
		MessageBody:       aws.String(" "), // must not me empty
	}

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
		addAttribute(entry, "name", namevalue)
	}
	if !valid {
		LogErr("Malformed command: %s", line)
		return
	}
	addAttribute(entry, "method", method)

	// Adding to batch
	smbi.Entries = append(smbi.Entries, entry)
	entryindex += 1

	// Sending if batch is full
	if entryindex == BatchSize {
		sendbatch()
		entryindex = 0
		smbi.Entries = smbi.Entries[:0]
	}
}

func sendbatch() {
	_, err := SqsClient.SendMessageBatch(Context, &smbi)
	PanicIfErr(err)
	Log("INF: Messages sent: %d", entryindex)
}
