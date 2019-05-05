package generalprobe

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/pkg/errors"
)

// SnsMessageAttributes is type to configure MessageAttributes of SNS.
type SnsMessageAttributes map[string]*sns.MessageAttributeValue

// PublishSnsScene is a scene to publish SNS message.
type PublishSnsScene struct {
	target  Target
	message []byte
	attrs   SnsMessageAttributes
	baseScene
}

// PublishSnsMessage creates a scene of SNS Publish with byte sequence message.
func (x *Generalprobe) PublishSnsMessage(target Target, message []byte) *PublishSnsScene {
	scene := PublishSnsScene{
		target:  target,
		message: message,
	}
	return &scene
}

// PublishSnsData creates a scene of SNS Publish with structure data.
func (x *Generalprobe) PublishSnsData(target Target, data interface{}) *PublishSnsScene {
	msg, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Fail to marshal data for SNS publish: %v", data)
	}

	return x.PublishSnsMessage(target, msg)
}

// MessageAttributes sets attribute of SNS MessageAttributes map
func (x *PublishSnsScene) MessageAttributes(attrs SnsMessageAttributes) *PublishSnsScene {
	x.attrs = attrs
	return x
}

// Strings return text explanation of the scene
func (x *PublishSnsScene) String() string {
	return fmt.Sprintf("SNS message to %s", x.target.arn())
}

func (x *PublishSnsScene) play() error {
	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(x.region()),
	}))
	snsService := sns.New(ssn)

	topicArn := x.target.arn()
	resp, err := snsService.Publish(&sns.PublishInput{
		Message:           aws.String(string(x.message)),
		TopicArn:          aws.String(topicArn),
		MessageAttributes: x.attrs,
	})

	logger.WithField("result", resp).Debug("sns:Publish result")

	if err != nil {
		return errors.Wrap(err, "Fail to publish report")
	}

	return nil
}
