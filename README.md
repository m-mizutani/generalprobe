# generalprobe

`generalprobe` is a framework of AWS CloudFormation integration test. Originally, "generalprobe" is a rehearsal which the actors and actresses who dress up in the same costumes on the same stage with the actual performance. This tool supports not unit testing but integration testing with deployed CloudFormation stack.

## Usage

```go
import (
	"testing"

	gp "github.com/m-mizutani/generalprobe"
	"github.com/stretchr/testify/assert"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
)

func TestSnsToDynamo(t *testing.T) {
	testID := uuid.New().String()

	g := gp.New("ap-northeast-1", "your-stack-name")
	g.AddScenes([]gp.Scene{
		// Send request
		gp.PublishSnsMessage("TriggerSNS", []byte(`{"id":"`+testID+`"}`)),

		// Recv result
		gp.GetDynamoRecord("ResultDB", func(table dynamo.Table) bool {
			var resp []map[string]interface{}
			err := table.Get("result_id", id).All(&resp)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(resp))
			assert.Equal(t, id, resp[0]["result_id"].(string))
			return true
		}),
	})

	g.Act()
}
```

## Available "Scenes"

- Publish SNS message
- Invoke Lambda function
- Read DynamoDB record
- Read Kinesis Stream record
