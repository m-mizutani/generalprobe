package generalprobe

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/pkg/errors"
)

type publishSnsMessage struct {
	target  Target
	message []byte
	attrs   SnsMessageAttributes
	baseScene
}

func (x *Generalprobe) PublishSnsMessage(target Target, message []byte) *publishSnsMessage {
	scene := publishSnsMessage{
		target:  target,
		message: message,
	}
	return &scene
}

type SnsMessageAttributes map[string]*sns.MessageAttributeValue

func PublishSnsMessageWithAttributes(target Target, message []byte, attrs SnsMessageAttributes) *publishSnsMessage {
	scene := publishSnsMessage{
		target:  target,
		message: message,
		attrs:   attrs,
	}
	return &scene
}

func (x *publishSnsMessage) play() error {
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
