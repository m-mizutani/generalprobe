package generalprobe_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	"github.com/stretchr/testify/require"

	gp "github.com/m-mizutani/generalprobe"
)

func TestSimplePlayBook(t *testing.T) {
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
			require.Equal(t, "ok", response.Message)
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
