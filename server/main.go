package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	. "github.com/cbytensky/bloxroute/common"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
)

type elem struct {
	order uint64
	value *string
}

func main() {
	CommonInit()

	mutex := sync.RWMutex{}
	var counter uint64
	storage := make(map[string]elem)

	rmi := &sqs.ReceiveMessageInput{
		MessageAttributeNames: []string{string(types.QueueAttributeNameAll)},
		QueueUrl:              &QueueUrl,
		MaxNumberOfMessages:   10,
	}

	nthreads, err := strconv.Atoi(os.Getenv("NTHREADS"))
	if err != nil {
		nthreads = runtime.NumCPU()
	}
	fmt.Println("Processing threads:", nthreads)

	messageChan := make(chan types.Message)

	for i := 0; i < nthreads; i++ {
		go func() {
			for {
				message := <-messageChan
				attributes := message.MessageAttributes
				println(*attributes["method"].StringValue)
				name := attributes["name"].StringValue
				if name != nil {
					println("name:", *name)
				}
				switch *attributes["method"].StringValue {

				case "AddItem":
					mutex.Lock()
					if _, exists := storage[*name]; exists {
						fmt.Printf("WRN: Overwriting: %s\n", *name)
					}
					counter += 1
					storage[*name] = elem{counter, message.Body}
					mutex.Unlock()

				case "RemoveItem":
					mutex.Lock()
					if _, exists := storage[*name]; !exists {
						LogErr("Not found: %s", *name)
					}
					delete(storage, *name)
					mutex.Unlock()

				case "GetItem":
					mutex.RLock()
					elem, exists := storage[*name]
					if exists {
						fmt.Printf("%s\n", *elem.value)
					} else {
						LogErr("Not found: %s", *name)
					}
					mutex.RUnlock()

				case "GetAllItems":
					mutex.RLock()
					storageSlice := make([]elem, 0, len(storage))
					for _, elem := range storage {
						storageSlice = append(storageSlice, elem)
					}
					sort.Slice(storageSlice, func(i, j int) bool {
						return storageSlice[i].order < storageSlice[j].order
					})
					for _, elem := range storageSlice {
						fmt.Print(*elem.value + " ")
					}
					fmt.Print("\n")
					mutex.RUnlock()

				}
			}
		}()
	}

	dmbi := sqs.DeleteMessageBatchInput{
		QueueUrl: &QueueUrl,
		Entries:  make([]types.DeleteMessageBatchRequestEntry, 0),
	}

	for {
		output, err := SqsClient.ReceiveMessage(Context, rmi)
		if err != nil {
			LogErr("%v", err)
		} else {
			numRecieved := len(output.Messages)
			if numRecieved > 0 {
				fmt.Printf("Recieved: %d\n", numRecieved)
				for i, message := range output.Messages {
					messageChan <- message
					dmbi.Entries = append(dmbi.Entries, types.DeleteMessageBatchRequestEntry{
						Id:            aws.String(string('0' + i)),
						ReceiptHandle: message.ReceiptHandle,
					})
				}
				_, err := SqsClient.DeleteMessageBatch(Context, &dmbi)
				PanicIfErr(err)
				dmbi.Entries = dmbi.Entries[:0]
			}
		}
	}
}
