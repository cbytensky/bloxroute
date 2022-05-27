package main

import (
	. "github.com/cbytensky/bloxroute/common"

	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// Type for data store entry
type elem struct {
	value *string
	order uint64 // Order of adding, used by GetAllItems
}

// The storage itself
var storage = make(map[string]elem)

// Mutex for concurrent access to the storage
var mutex = sync.RWMutex{}

// Counter for storing as order for AddItem
var counter uint64

// Type for slice entry for sorting for GetAllItems
type sortElem struct{order uint64; name string; value *string}

func main() {
	CommonInit()


	rmi := &sqs.ReceiveMessageInput{ // reused for ReceiveMessage call
		MessageAttributeNames: []string{string(types.QueueAttributeNameAll)}, // We want to receive all attributes
		QueueUrl:              &QueueUrl,
		MaxNumberOfMessages:   BatchSize,
		WaitTimeSeconds:       20,
	}

	// Determing number of goroutines
	nthreads, err := strconv.Atoi(os.Getenv("NTHREADS"))
	if err != nil {
		nthreads = runtime.NumCPU() // fallback to number of CPUs
	}
	Log("INF: Processing threads: %d", nthreads)

	// Channel for passing received messages from main thread to processing goroutines
	messageChan := make(chan types.Message, BatchSize) // BatchSize for challen length seems to be good enough

	// Launching processing goroutines
	for i := 0; i < nthreads; i++ {
		go func() { // Processing goroutine
			for {
				message := <-messageChan
				attributes := message.MessageAttributes
				name := attributes["name"].StringValue

				switch *attributes["method"].StringValue {

				case "AddItem":
					mutex.Lock()
					if _, exists := storage[*name]; exists {
						Log("WRN: Overwriting: %s", *name)
					}
					storage[*name] = elem{message.Body, counter}
					counter += 1
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
						Log("%s", *elem.value)
					} else {
						LogErr("Not found: %s", *name)
					}
					mutex.RUnlock()

				case "GetAllItems":
					mutex.RLock()
					// We need to sort all items by `order`
					storageSlice := make([]*sortElem , 0, len(storage)) // Slice for sorting
					// Copying data from storage map to slice
					for name, elem := range storage {
						storageSlice = append(storageSlice, &sortElem{ elem.order, name, elem.value })
					}
					//Sorting
					sort.Slice(storageSlice, func(i, j int) bool {
						return storageSlice[i].order < storageSlice[j].order
					})
					// Printing
					for _, e := range storageSlice {
						Log("%s: %s", e.name, *e.value)
					}
					mutex.RUnlock()

				}
			}
		}()
	}

	dmbi := sqs.DeleteMessageBatchInput{ // reused in DeleteMessageBatch
		QueueUrl: &QueueUrl,
		Entries:  make([]types.DeleteMessageBatchRequestEntry, 0),
	}

	// Main loop: receiving messages
	for {
		output, err := SqsClient.ReceiveMessage(Context, rmi)
		if err != nil {
			LogErr("%v", err)
		} else {
			numRecieved := len(output.Messages)
			if numRecieved > 0 {
				Log("INF: Messages received: %d", numRecieved)
				for i, message := range output.Messages {
					// Sending message to processing goroutines
					messageChan <- message
					// Adding message receipt to DeleteMessageBatch
					dmbi.Entries = append(dmbi.Entries, types.DeleteMessageBatchRequestEntry{
						Id:            DigitToStr(uint8(i)),
						ReceiptHandle: message.ReceiptHandle,
					})
				}
				// Deleting received messages
				_, err := SqsClient.DeleteMessageBatch(Context, &dmbi)
				PanicIfErr(err)
				dmbi.Entries = dmbi.Entries[:0]
			}
		}
	}
}
