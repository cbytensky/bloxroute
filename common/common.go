package common
import (
	"os"
	"fmt"
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var QueueUrl = os.Getenv("QUEUE_URL")
var SqsClient *sqs.Client
var Context = context.TODO()

func CommonInit() {
	awsConfig, err := config.LoadDefaultConfig(Context)
	PanicIfErr(err)
	SqsClient = sqs.NewFromConfig(awsConfig)
}

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func LogErr(format string, args ...interface{}) {
	fmt.Printf("ERR: " + format + "\n", args...)
}
