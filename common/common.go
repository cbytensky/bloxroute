package common
import (
	"os"
	"fmt"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Batch size for SQS API calls
const BatchSize = 10

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

func DigitToStr(digit int) *string {
	return aws.String(string('0' + uint8(digit)))
}

func Log(format string, args ...interface{}) {
	fmt.Printf(format + "\n", args...)
}

func LogErr(format string, args ...interface{}) {
	Log("ERR: " + format, args...)
}
