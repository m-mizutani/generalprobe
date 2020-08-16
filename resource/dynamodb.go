package resource

import "github.com/aws/aws-sdk-go/aws/session"

// DynamoDB is resource entity for AWS DynamoDB.
type DynamoDB struct {
	// AWS SDK session to access lambda function
	Session *session.Session
	// TableName is DynamoDB table name.
	TableName string
}
