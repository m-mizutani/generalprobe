package generalprobe

import (
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/pkg/errors"
)

type SnsMessageAttributes map[string]*sns.MessageAttributeValue

type PublishSns struct {
	target  Target
	message []byte
	attrs   SnsMessageAttributes
	baseScene
}

func (x *Generalprobe) PublishSnsMessage(target Target, message []byte) *PublishSns {
	scene := PublishSns{
		target:  target,
		message: message,
	}
	return &scene
}

func (x *Generalprobe) PublishSnsData(target Target, data interface{}) *PublishSns {
	msg, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Fail to marshal data for SNS publish: %v", data)
	}

	return x.PublishSnsMessage(target, msg)
}

func (x *PublishSns) AddMessageAttributes(attrs SnsMessageAttributes) *PublishSns {
	x.attrs = attrs
	return x
}

func (x *PublishSns) play() error {
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
