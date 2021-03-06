package generalprobe

import (
	"crypto/sha256"
	"fmt"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

// PutKinesisStreamRecordScene is a scene to put a new Kinesis record.
type PutKinesisStreamRecordScene struct {
	target Target
	baseScene
	message []byte
}

// PutKinesisStreamRecord is a constructor of Scene
func PutKinesisStreamRecord(target Target, message []byte) *PutKinesisStreamRecordScene {
	scene := PutKinesisStreamRecordScene{
		target:  target,
		message: message,
	}
	return &scene
}

// Strings return text explanation of the scene
func (x *PutKinesisStreamRecordScene) string() string {
	return fmt.Sprintf("Put a new kinesis record to %s", x.target.arn(x.gp))
}

func (x *PutKinesisStreamRecordScene) play() error {
	const maxRetry = 20

	streamName := x.target.name(x.gp)

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(x.region()),
	}))
	kinesisService := kinesis.New(ssn)

	kinesisInput := kinesis.PutRecordInput{
		Data:         x.message,
		PartitionKey: aws.String(fmt.Sprintf("%x", sha256.Sum256(x.message))),
		StreamName:   aws.String(streamName),
	}
	resp, err := kinesisService.PutRecord(&kinesisInput)

	logger.WithField("resp", resp).Debug("Done Kinesis PutRecord")
	if err != nil {
		return errors.Wrap(err, "Fail to put kinesis record")
	}

	return nil
}
