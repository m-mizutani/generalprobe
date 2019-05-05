package generalprobe_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	gp "github.com/m-mizutani/generalprobe"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var logger = logrus.New()

func init() {
	logLevel := logrus.WarnLevel
	switch os.Getenv("GP_LOG_LEVEL") {
	case "TRACE":
		logLevel = logrus.TraceLevel
	case "DEBUG":
		logLevel = logrus.DebugLevel
	case "INFO":
		logLevel = logrus.InfoLevel
	case "WARN":
		logLevel = logrus.WarnLevel
	case "ERROR":
		logLevel = logrus.ErrorLevel

	}
	logger.SetLevel(logLevel)
}

type testParameters struct {
	StackName string `json:"StackName"`
	Region    string `json:"Region"`
}

func loadTestParameters() testParameters {
	paramFile := "test-stack/params.json"
	fd, err := os.Open(paramFile)
	if err != nil {
		logger.Printf("Can not open")
		logger.Error(err)
	}

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		logger.Error(err)
	}

	var p testParameters
	err = json.Unmarshal(data, &p)
	if err != nil {
		logger.Error(err)
	}

	return p
}

func TestBasicUsage(t *testing.T) {
	params := loadTestParameters()
	g := gp.New(params.Region, params.StackName)

	n := 0
	g.AddScenes([]gp.Scene{
		g.AdLib(func() {
			n++
		}),
		g.Pause(1),
		g.AdLib(func() {
			n++
		}),
	})
	g.Run()
	assert.Equal(t, 2, n)
}

func TestSnsToDynamo(t *testing.T) {
	params := loadTestParameters()
	id := uuid.New().String()

	done := false
	g := gp.New(params.Region, params.StackName)
	g.AddScenes([]gp.Scene{
		// Send request
		g.PublishSnsMessage(g.LogicalID("Trigger"), []byte(`{"id":"`+id+`"}`)),

		// Recv result
		g.GetDynamoRecord(g.LogicalID("ResultStore"), func(table dynamo.Table) bool {
			var resp []map[string]interface{}
			err := table.Get("result_id", id).All(&resp)
			logger.WithField("dynamo resp", resp).Debug("get dynamo response")
			require.NoError(t, err)
			assert.Equal(t, 1, len(resp))
			assert.Equal(t, id, resp[0]["result_id"].(string))
			done = true
			return true
		}),

		g.GetLambdaLogs(g.LogicalID("TestHandler"), func(logs gp.CloudWatchLog) bool {
			assert.True(t, logs.Contains(id))
			return true
		}).Filter(id),
	})

	g.Run()
	require.Equal(t, true, done)
}

func TestKinesisStream(t *testing.T) {
	params := loadTestParameters()

	id := uuid.New().String()
	g := gp.New(params.Region, params.StackName)
	g.AddScenes([]gp.Scene{
		// Send message
		g.PutKinesisStreamRecord(g.LogicalID("ResultStream"), []byte(id)),
		g.GetKinesisStreamRecord(g.LogicalID("ResultStream"), func(data []byte) bool {
			assert.Equal(t, string(data), id)
			return true
		}),
	})

	g.Run()

}

func TestSearchLambdaLogsNotFound(t *testing.T) {
	params := loadTestParameters()
	g := gp.New(params.Region, params.StackName)

	logs := g.SearchLambdaLogs(gp.SearchLambdaLogsArgs{
		LambdaTarget: g.Arn("arn:aws:lambda:ap-northeast-1:1234567890:function:no-such-function"),
		QueryLimit:   1,
	})

	assert.Equal(t, 0, len(logs))
}
