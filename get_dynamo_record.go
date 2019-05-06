package generalprobe

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/pkg/errors"
)

// GetDynamoRecordScene is a scene of waiting DynamoDB data.
type GetDynamoRecordScene struct {
	target Target

	callback GetDynamoRecordCallback
	pollingScene
}

// GetDynamoRecordCallback is callback function called after retrieving target record
type GetDynamoRecordCallback func(table dynamo.Table) bool

// GetDynamoRecord is a constructor of Scene
func GetDynamoRecord(target Target, callback GetDynamoRecordCallback) *GetDynamoRecordScene {
	scene := GetDynamoRecordScene{
		target:   target,
		callback: callback,
		pollingScene: pollingScene{
			limit:    20,
			interval: 3,
		},
	}
	return &scene
}

// Strings return text explanation of the scene
func (x *GetDynamoRecordScene) string() string {
	return fmt.Sprintf("Read DynamoDB of %s", x.target.arn(x.gp))
}

func (x *GetDynamoRecordScene) play() error {
	db := dynamo.New(session.New(), &aws.Config{
		Region: aws.String(x.region()),
	})
	table := db.Table(x.target.name(x.gp))

	for n := 0; n < x.limit; n++ {
		if x.callback(table) {
			return nil
		}

		time.Sleep(time.Second * time.Duration(x.interval))
	}

	return errors.New("Timeout to fetch records from DynamoDB")
}
