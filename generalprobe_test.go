package generalprobe_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	gp "github.com/m-mizutani/generalprobe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testParameters struct {
	StackName string `json:"StackName"`
	Region    string `json:"Region"`
}

func loadTestParameters() testParameters {
	paramFile := "test-stack/params.json"
	fd, err := os.Open(paramFile)
	if err != nil {
		log.Fatalf("Can not open %s: %s", paramFile, err)
	}

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Fatalf("Fail to read data %s: %s", paramFile, err)
	}

	var p testParameters
	err = json.Unmarshal(data, &p)
	if err != nil {
		log.Fatalf("Fail to unmarshal data %s: %s", paramFile, err)
	}

	return p
}

func TestBasicUsage(t *testing.T) {
	params := loadTestParameters()

	n := 0
	scenario := []gp.Scene{
		gp.AdLib(func() {
			n++
		}),
		gp.Pause(1),
		gp.AdLib(func() {
			n++
		}),
	}

	err := gp.New(params.Region, params.StackName).Play(scenario)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
}

func TestSnsToDynamo(t *testing.T) {
	params := loadTestParameters()
	id := uuid.New().String()

	done := false
	scenario := []gp.Scene{
		// Send request
		gp.PublishSnsMessage(gp.LogicalID("Trigger"), []byte(`{"id":"`+id+`"}`)),

		// Recv result
		gp.GetDynamoRecord(gp.LogicalID("ResultStore"), func(table dynamo.Table) bool {
			var resp []map[string]interface{}
			err := table.Get("result_id", id).All(&resp)

			require.NoError(t, err)
			if len(resp) > 0 {
				assert.Equal(t, 1, len(resp))
				assert.Equal(t, id, resp[0]["result_id"].(string))
				done = true
				return true
			}

			return false
		}),

		gp.GetLambdaLogs(gp.LogicalID("TestHandler"), func(logs gp.CloudWatchLog) bool {
			assert.True(t, logs.Contains(id))
			return true
		}).Filter(id),
	}

	err := gp.New(params.Region, params.StackName).Play(scenario)
	require.NoError(t, err)
	require.Equal(t, true, done)
}

func TestInvokeLambda(t *testing.T) {
	params := loadTestParameters()

	id := uuid.New().String()
	request := struct {
		ID string `json:"id"`
	}{ID: id}
	response := struct {
		Message string `json:"message"`
	}{}

	scenario := []gp.Scene{
		// Invoke Lambda with SNS event as argument
		gp.InvokeLambda(gp.LogicalID("TestHandler"), func(ret []byte) {
			err := json.Unmarshal(ret, &response)
			require.NoError(t, err)
			assert.Equal(t, "ok", response.Message)
		}).SnsEvent(request),

		// Read result from DynamoDB that TestHandler wrote
		gp.GetDynamoRecord(gp.LogicalID("ResultStore"), func(table dynamo.Table) bool {
			var resp []map[string]interface{}
			err := table.Get("result_id", id).All(&resp)
			require.NoError(t, err)

			// If returning true, the scene will exit and move next scene.
			// If returning false, the scene continues polling and the callback will be
			// invoked after 3 second in default. This polling retries up to 20 times in default.
			return len(resp) > 0
		}),
	}

	err := gp.New(params.Region, params.StackName).Play(scenario)
	require.NoError(t, err)
}

func TestKinesisStream(t *testing.T) {
	params := loadTestParameters()

	id := uuid.New().String()
	scenario := []gp.Scene{
		// Send message
		gp.PutKinesisStreamRecord(gp.LogicalID("ResultStream"), []byte(id)),
		gp.GetKinesisStreamRecord(gp.LogicalID("ResultStream"), func(data []byte) bool {
			assert.Equal(t, string(data), id)
			return true
		}),
	}

	err := gp.New(params.Region, params.StackName).Play(scenario)
	require.NoError(t, err)
}
