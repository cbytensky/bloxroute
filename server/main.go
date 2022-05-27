package main

import (
	. "common"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"sort"
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

	for {
		output, err := SqsClient.ReceiveMessage(Context, rmi)
		if err != nil {
			LogErr("%v", err)
		} else {
			numRecieved := len(output.Messages)
			if numRecieved > 0 {
				fmt.Printf("Recieved: %d\n", numRecieved)
				for _, message := range output.Messages {
					go func(message types.Message) {
						attributes := message.MessageAttributes
						name := attributes["name"].StringValue
						println(*attributes["method"].StringValue)
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
						SqsClient.DeleteMessage(Context, &sqs.DeleteMessageInput{QueueUrl: &QueueUrl, ReceiptHandle: message.ReceiptHandle})
					}(message)
				}
			}
		}
	}
}
