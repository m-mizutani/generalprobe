# generalprobe

`generalprobe` is a framework of AWS CloudFormation integration test. Originally, "generalprobe" is a rehearsal which the actors and actresses who dress up in the same costumes on the same stage with the actual performance. This tool supports not unit testing but integration testing with deployed CloudFormation stack.

## Usage

```go
import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/guregu/dynamo"

	gp "github.com/m-mizutani/AlertResponder/generalprobe"
)


func TestYourStack(t *testing.T) {
	playbook := gp.NewPlaybook("ap-northeast-1", "your-stack-name")
	playbook.SetScenario([]gp.Scene{
		gp.PublishSnsMessage("Notification", "xyz"),
		gp.Pause(3),
		gp.GetDynamoRecord("NotificationLog", func(table dynamo.Table) (bool, error) {
		    var result struct {
			    Key string `dynamo:"key"`
			}
			err := table.Get("key", "xyz").One(&result)
			require.NoError(t, err)
			assert.Equal(t, "xyz", result.Key)
			return true, nil
		}
	})
	playbook.Act()
}
```
