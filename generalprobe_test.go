package generalprobe_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	gp "github.com/m-mizutani/generalprobe"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
)

func init() {
	log.SetLevel(log.InfoLevel)
}

type testParameters struct {
	StackName    string `json:"StackName"`
	Region       string `json:"Region"`
	SnsName      string `json:"SnsName"`
	LambdaName   string `json:"LambdaName"`
	DynamoDBName string `json:"DynamoDBName"`
	S3BucketName string `json:"S3BucketName"`
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

	type respMsg struct {
		ID string `dynamo:"result_id"`
		Report []byte `dynamo:"report"`
	}

	done := false
	g := gp.New(params.Region, params.StackName)
	g.AddScenes([]gp.Scene{
		// Send request
		gp.PublishSnsMessage("Trigger", []byte(`{"id":"`+id+`"}`)),

		// Recv result
		gp.GetDynamoRecord("ResultStore", func(table dynamo.Table) bool {
			var resp []respMsg
			err := table.Get("result_id", id).All(&resp)
			log.WithField("dynamo resp", resp).Debug("get dynamo response")
			require.NoError(t, err)
			assert.Equal(t, 1, len(resp))
			assert.Equal(t, id, resp[0].ID)
			done = true
			return true
		}),
	})

	g.Act()
	require.Equal(t, true, done)
}

