package generalprobe

import (
	"crypto/sha256"
	"fmt"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	log "github.com/sirupsen/logrus"
)

type putKinesisStreamRecord struct {
	target Target
	baseScene
	message []byte
}

// PutKinesisStreamRecord is a constructor of Scene
func PutKinesisStreamRecord(target Target, message []byte) *putKinesisStreamRecord {
	scene := putKinesisStreamRecord{
		target:  target,
		message: message,
	}
	return &scene
}

func (x *putKinesisStreamRecord) play() error {
	const maxRetry = 20

	streamName := x.target.name()

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

	log.WithField("resp", resp).Debug("Done Kinesis PutRecord")
	if err != nil {
		return errors.Wrap(err, "Fail to put kinesis record")
	}

	return nil
}
