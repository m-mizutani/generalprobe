package generalprobe

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

// GetKinesisStreamRecordScene is a scene of waiting Kinesis record.
type GetKinesisStreamRecordScene struct {
	target Target
	pollingScene
	callback GetKinesisStreamRecordCallback
}

// GetKinesisStreamRecordCallback is callback function called after retrieving kinesis record
type GetKinesisStreamRecordCallback func(data []byte) bool

// GetKinesisStreamRecord is a constructor of Scene
func GetKinesisStreamRecord(target Target, callback GetKinesisStreamRecordCallback) *GetKinesisStreamRecordScene {
	scene := GetKinesisStreamRecordScene{
		target:   target,
		callback: callback,
		pollingScene: pollingScene{
			limit:    20,
			interval: 3,
		},
	}
	return &scene
}

func (x *GetKinesisStreamRecordScene) string() string {
	return fmt.Sprintf("Get Kinesis Record from %s", x.target.arn(x.gp))
}

func (x *GetKinesisStreamRecordScene) play() error {
	streamName := x.target.name(x.gp)

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(x.region()),
	}))

	kinesisService := kinesis.New(ssn)
	resp, err := kinesisService.ListShards(&kinesis.ListShardsInput{
		StreamName: aws.String(streamName),
	})
	if err != nil {
		logger.Fatal("Fail to shard list", err)
	}

	shardList := []string{}
	for _, s := range resp.Shards {
		shardList = append(shardList, *s.ShardId)
	}

	if len(shardList) != 1 {
		logger.Fatal("Invalid shard number: ", len(shardList), ", expected 1")
	}

	now := time.Now()

	iter, err := kinesisService.GetShardIterator(&kinesis.GetShardIteratorInput{
		ShardId:           aws.String(shardList[0]),
		ShardIteratorType: aws.String("AT_TIMESTAMP"),
		StreamName:        aws.String(streamName),
		Timestamp:         &now,
	})
	if err != nil {
		logger.Fatal("Fail to get iterator", err)
	}

	shardIter := iter.ShardIterator
	for i := 0; i < x.limit; i++ {
		records, err := kinesisService.GetRecords(&kinesis.GetRecordsInput{
			ShardIterator: shardIter,
		})

		if err != nil {
			logger.WithField("records", records).Fatal("Fail to get kinesis records")
		}
		shardIter = records.NextShardIterator

		if len(records.Records) > 0 {
			for _, record := range records.Records {
				if x.callback(record.Data) {
					return nil
				}
			}
		}

		time.Sleep(time.Second * time.Duration(x.interval))
	}

	return errors.New("No kinesis message")
}
