package generalprobe

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/pkg/errors"
)

type getDynamoRecord struct {
	target Target

	callback GetDynamoRecordCallback
	baseScene
}

// GetDynamoRecordCallback is callback function called after retrieving target record
type GetDynamoRecordCallback func(table dynamo.Table) bool

// GetDynamoRecord is a constructor of Scene
func (x *Generalprobe) GetDynamoRecord(target Target, callback GetDynamoRecordCallback) *getDynamoRecord {
	scene := getDynamoRecord{
		target:   target,
		callback: callback,
	}
	return &scene
}

func (x *getDynamoRecord) play() error {
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(x.region())})
	table := db.Table(x.target.name())
	const maxRetry int = 30

	for n := 0; n < maxRetry; n++ {
		time.Sleep(time.Second * 2)

		if x.callback(table) {
			return nil
		}
	}

	return errors.New("Timeout to fetch records from DynamoDB")
}
