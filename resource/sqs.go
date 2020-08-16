package resource

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// SQS is resource entity for AWS SQS.
type SQS struct {
	// AWS SDK session to access lambda function
	Session *session.Session
	// QueueURL is SQS topic ARN.
	QueueURL string
}

func (x *SQS) SendMessage(body string) *SendSQSScene {
	return &SendSQSScene{
		sqs:  x,
		body: body,
	}
}

func (x *SQS) SendObject(obj interface{}) *SendSQSScene {
	raw, err := json.Marshal(obj)
	if err != nil {
		log.Fatalf("Failed to json.Marshal of SendObject: %v, %+v", obj, err)
	}

	return &SendSQSScene{
		sqs:  x,
		body: string(raw),
	}
}

type SendSQSScene struct {
	sqs      *SQS
	body     string
	callback func(messageID string)
	input    sqs.SendMessageInput
}

func (x *SendSQSScene) SetInput(input *sqs.SendMessageInput) *SendSQSScene {
	x.input = *input
	return x
}

func (x *SendSQSScene) Name() string {
	return fmt.Sprintf("Publish message to %s", x.sqs.QueueURL)
}

func (x *SendSQSScene) Play() (bool, error) {
	client := sqs.New(x.sqs.Session)
	output, err := client.SendMessage(&x.input)
	if err != nil {
		return true, err
	}
	if x.callback != nil {
		x.callback(aws.StringValue(output.MessageId))
	}

	return true, nil
}
