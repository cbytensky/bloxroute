package main

import (
	. "common"

	"fmt"
	"os"
	"bufio"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

var smbi = sqs.SendMessageBatchInput{ QueueUrl: &QueueUrl }
var entryindex uint8

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
	if stat.Mode() & os.ModeCharDevice == 0 { // Checking stdin is not TTY
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

var entries = make([]types.SendMessageBatchRequestEntry, 0)

func addAttribute(entry types.SendMessageBatchRequestEntry, name string, value string) {
	entry.MessageAttributes[name] = types.MessageAttributeValue{
		DataType: aws.String("String"),
		StringValue: &value,
	}
}

func addtobatch(line string) {

	entry := types.SendMessageBatchRequestEntry{
		Id: aws.String(string('0' + entryindex)), // Just digits '0'..'9'
		MessageAttributes: make(map[string]types.MessageAttributeValue),
		MessageBody: aws.String("empty"),
	}

	command := line[0]
	valid := true
	method := "GetAllItems"
	if len(line) == 1 {
		valid = valid && command == '!'
	} else {
		namevalue := line[1:]
		if command == '+' {
			method = "AddItem"
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
	entries = append(entries, entry)
	entryindex += 1

	// Sending if batch is full
	if entryindex == 10 {
		sendbatch()
		entryindex = 0
		entries = entries[:0]
	}
}

func sendbatch(){
	smbi.Entries = entries
	_, err := SqsClient.SendMessageBatch(Context, &smbi)
	PanicIfErr(err)
	fmt.Printf("Messages sent: %d\n", entryindex)
}
