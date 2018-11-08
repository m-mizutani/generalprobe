package generalprobe

import (
	"time"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	log "github.com/sirupsen/logrus"
)

type getKinesisStreamRecord struct {
	logicalID string
	baseScene
	callback GetKinesisStreamRecordCallback
}

// GetKinesisStreamRecordCallback is callback function called after retrieving kinesis record
type GetKinesisStreamRecordCallback func(data []byte) bool

// GetKinesisStreamRecord is a constructor of Scene
func GetKinesisStreamRecord(logicalID string, callback GetKinesisStreamRecordCallback) *getKinesisStreamRecord {
	scene := getKinesisStreamRecord{
		logicalID: logicalID,
		callback:  callback,
	}
	return &scene
}

func (x *getKinesisStreamRecord) play() error {
	const maxRetry = 20

	streamName := x.lookupPhysicalID(x.logicalID)

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(x.region()),
	}))

	kinesisService := kinesis.New(ssn)
	resp, err := kinesisService.ListShards(&kinesis.ListShardsInput{
		StreamName: aws.String(streamName),
	})
	if err != nil {
		log.Fatal("Fail to shard list", err)
	}

	shardList := []string{}
	for _, s := range resp.Shards {
		shardList = append(shardList, *s.ShardId)
	}

	if len(shardList) != 1 {
		log.Fatal("Invalid shard number: ", len(shardList), ", expected 1")
	}

	now := time.Now()

	iter, err := kinesisService.GetShardIterator(&kinesis.GetShardIteratorInput{
		ShardId:           aws.String(shardList[0]),
		ShardIteratorType: aws.String("AT_TIMESTAMP"),
		StreamName:        aws.String(streamName),
		Timestamp:         &now,
	})
	if err != nil {
		log.Fatal("Fail to get iterator", err)
	}

	shardIter := iter.ShardIterator
	for i := 0; i < maxRetry; i++ {
		records, err := kinesisService.GetRecords(&kinesis.GetRecordsInput{
			ShardIterator: shardIter,
		})

		if err != nil {
			log.WithField("records", records).Fatal("Fail to get kinesis records")
		}
		shardIter = records.NextShardIterator

		if len(records.Records) > 0 {
			for _, record := range records.Records {
				if x.callback(record.Data) {
					return nil
				}
			}
		}

		time.Sleep(time.Second * 1)
	}

	return errors.New("No kinesis message")
}
