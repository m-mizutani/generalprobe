# Generalprobe

[![GoDoc](https://godoc.org/github.com/m-mizutani/generalprobe?status.svg)](https://godoc.org/github.com/m-mizutani/generalprobe)

`generalprobe` is an integration test framework for Serverless Application of AWS CloudFormation. A developer define `Scene` as event regarding the serverless application, then this framework invoke the defined events sequentially. The developper can inject test code into the scenes, e.g. Invoke AWS Lambda and check DynamoDB that the Lambda write some item into.

Originally, _generalprobe_ is a rehearsal which the actors and actresses who dress up in the same costumes on the same stage with the actual performance. It's called _dress rehearsal_ in English. See [wikipedia](https://de.wikipedia.org/wiki/Generalprobe) for more detail.

## Getting Started

A following sample shows a test of sample stack. In the scenario, [Lambda function](https://github.com/m-mizutani/generalprobe/blob/master/test-stack/main.py) is invoked at first phase and checking DynamoDB record is second phase.

```go
package generalprobe_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	"github.com/stretchr/testify/require"

	gp "github.com/m-mizutani/generalprobe"
)

func TestSimplePlayBook(t *testing.T) {
	id := uuid.New().String()
	request := struct {
		ID string `json:"id"`
	}{ID: id}
	response := struct {
		Message string `json:"message"`
	}{}

	playbook := []gp.Scene{
		// Invoke Lambda with SNS event as argument
		gp.InvokeLambda(gp.LogicalID("TestHandler"), func(ret []byte) {
			err := json.Unmarshal(ret, &response)
			require.NoError(t, err)
			require.Equal(t, "ok", response.Message)
		}).SnsEvent(request),

		// Read result from DynamoDB that TestHandler wrote
		gp.GetDynamoRecord(gp.LogicalID("ResultStore"), func(table dynamo.Table) bool {
			var resp []map[string]interface{}
			// Depending to github.com/guregu/dynamo regarding DynamoDB access
			err := table.Get("result_id", id).All(&resp)
			require.NoError(t, err)

			// If returning, the callback will be invoked again to poll
			return len(resp) > 0
		}),
	}

	err := gp.New(os.Getenv("TEST_REGION"), os.Getenv("TEST_STACKNAME")).Play(playbook)
	require.NoError(t, err)
}
```

## Scenes

`Scene` is unit of event on the serverless application in Generalprobe. `Play()` function of `generalprobe` executes scenes (slice of `Scene`) sequentially. Available scenes are following.

### Publish SNS message

```go
gp.PublishSnsMessage(gp.LogicalID("TopicName"), []byte(`{"id":"`+id+`"}`))
gp.PublishSnsData(gp.LogicalID("TopicName"), yourObject))
```

See also

- [PublishSnsData](https://godoc.org/github.com/m-mizutani/generalprobe#PublishSnsData)
- [PublishSnsMessage](https://godoc.org/github.com/m-mizutani/generalprobe#PublishSnsMessage)

### Invoke Lambda function

```go
gp.InvokeLambda(gp.LogicalID("FuncName"), func(ret []byte) {
	err := json.Unmarshal(ret, &response)
	require.NoError(t, err)
	require.Equal(t, "ok", response.Message)
}).SnsEvent(request)
```

`InvokeLambda` requires not only invoke target but alos callback to receive a result of Lambda. If you do not need to check a result, nothing to do in the callback.

See also [InvokeLambda](https://godoc.org/github.com/m-mizutani/generalprobe#InvokeLambda)

### Read Lambda logs from CloudWatch Logs

```go
gp.GetLambdaLogs(gp.LogicalID("FuncName"), func(logs gp.CloudWatchLog) bool {
	assert.True(t, logs.Contains(id))
	return true
}).Filter(id),
```

See also [GetLambdaLogs](https://godoc.org/github.com/m-mizutani/generalprobe#GetLambdaLogs)

### Read DynamoDB record

```go
gp.GetDynamoRecord(gp.LogicalID("TableName"), func(table dynamo.Table) bool {
	var resp []map[string]interface{}
	err := table.Get("result_id", id).All(&resp)
	require.NoError(t, err)
	return len(resp) > 0
}),
```

See also [GetDynamoRecord](https://godoc.org/github.com/m-mizutani/generalprobe#GetDynamoRecord)

### Get Kinesis Stream record

```go
gp.GetKinesisStreamRecord(gp.LogicalID("StreamName"), func(data []byte) bool {
	assert.Equal(t, string(data), id)
	return true
}),
```

See also [GetKinesisStreamRecord](https://godoc.org/github.com/m-mizutani/generalprobe#GetKinesisStreamRecord)

### Put Kinesis Stream record

```go
gp.PutKinesisStreamRecord(gp.LogicalID("StreamName"), []byte(id)),
```

See also [PutKinesisStreamRecord](mizutani/generalprobe#PutKinesisStreamRecord)

## Target

To specify AWS resource. `LogicalID` specifies resource name of CloudFormation and convert the resource name to ARN. `Arn` specifies ARN and it should be used to refer resource that is not under management of CloudFormation stack.

See also
- [LogicalID](https://godoc.org/github.com/m-mizutani/generalprobe#LogicalID)
- [Arn](https://godoc.org/github.com/m-mizutani/generalprobe#Arn)