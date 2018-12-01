package generalprobe

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type publishSnsMessage struct {
	target  Target
	message []byte
	baseScene
}

func PublishSnsMessage(target Target, message []byte) *publishSnsMessage {
	scene := publishSnsMessage{
		target:  target,
		message: message,
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
		Message:  aws.String(string(x.message)),
		TopicArn: aws.String(topicArn),
	})

	log.WithField("result", resp).Debug("sns:Publish result")

	if err != nil {
		return errors.Wrap(err, "Fail to publish report")
	}

	return nil
}
