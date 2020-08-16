package resource

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/pkg/errors"
)

// SNS is resource entity for AWS SNS.
type SNS struct {
	// AWS SDK session to access lambda function
	Session *session.Session
	// TopicARN is SNS topic ARN.
	TopicARN string
}

type PublishSNSScene struct {
	sns      *SNS
	message  string
	input    sns.PublishInput
	callback func(msgID string)
}

func (x *SNS) PublishMessage(msg string) *PublishSNSScene {
	return &PublishSNSScene{
		sns:     x,
		message: msg,
	}
}

func (x *SNS) PublishObject(obj interface{}) *PublishSNSScene {
	raw, err := json.Marshal(obj)
	if err != nil {
		log.Fatalf("Failed to json.Marshal in PublishObject: %v, %+v", obj, err)
	}

	return &PublishSNSScene{
		sns:     x,
		message: string(raw),
	}
}

func (x *PublishSNSScene) Callback(f func(string)) {
	x.callback = f
}

// SetInput copies original sns.PublishInput to own. TopicARN and Message will be replaced if available.
func (x *PublishSNSScene) SetInput(input *sns.PublishInput) *PublishSNSScene {
	x.input = *input
	return x
}

func (x *PublishSNSScene) Name() string {
	return fmt.Sprintf("Pubslih SNS to %s", x.sns.TopicARN)
}

func (x *PublishSNSScene) Play() (bool, error) {
	client := sns.New(x.sns.Session)
	x.input.TopicArn = aws.String(x.sns.TopicARN)
	if x.message == "" {
		x.input.Message = aws.String(x.message)
	}

	output, err := client.Publish(&x.input)
	if err != nil {
		return true, errors.Wrap(err, "Failed to publish SNS")
	}

	if x.callback != nil {
		x.callback(aws.StringValue(output.MessageId))
	}

	return true, nil
}
