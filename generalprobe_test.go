package generalprobe_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	gp "github.com/m-mizutani/generalprobe"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	logLevel := log.WarnLevel
	switch os.Getenv("GP_LOG_LEVEL") {
	case "TRACE":
		logLevel = log.TraceLevel
	case "DEBUG":
		logLevel = log.DebugLevel
	case "INFO":
		logLevel = log.InfoLevel
	case "WARN":
		logLevel = log.WarnLevel
	case "ERROR":
		logLevel = log.ErrorLevel

	}
	log.SetLevel(logLevel)
}

type testParameters struct {
	StackName    string `json:"StackName"`
	Region       string `json:"Region"`
}

func loadTestParameters() testParameters {
	paramFile := "test-stack/params.json"
	fd, err := os.Open(paramFile)
	if err != nil {
		log.Printf("Can not open")
		log.Error(err)
	}

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Error(err)
	}

	var p testParameters
	err = json.Unmarshal(data, &p)
	if err != nil {
		log.Error(err)
	}

	return p
}

func TestBasicUsage(t *testing.T) {
	params := loadTestParameters()
	g := gp.New(params.Region, params.StackName)

	n := 0
	g.AddScenes([]gp.Scene{
		gp.AdLib(func() {
			n++
		}),
		gp.Pause(1),
		gp.AdLib(func() {
			n++
		}),
	})
	g.Act()
	assert.Equal(t, 2, n)
}

func TestSnsToDynamo(t *testing.T) {
	params := loadTestParameters()
	id := uuid.New().String()

	done := false
	g := gp.New(params.Region, params.StackName)
	g.AddScenes([]gp.Scene{
		// Send request
		gp.PublishSnsMessage("Trigger", []byte(`{"id":"`+id+`"}`)),

		// Recv result
		gp.GetDynamoRecord("ResultStore", func(table dynamo.Table) bool {
			var resp []map[string]interface{}
			err := table.Get("result_id", id).All(&resp)
			log.WithField("dynamo resp", resp).Debug("get dynamo response")
			require.NoError(t, err)
			assert.Equal(t, 1, len(resp))
			assert.Equal(t, id, resp[0]["result_id"].(string))
			done = true
			return true
		}),

		gp.AdLib(func() {
			logs := g.SearchLambdaLogs("TestHandler", id)
			assert.NotEqual(t, 0, len(logs))
		}),
	})

	g.Act()
	require.Equal(t, true, done)
}
