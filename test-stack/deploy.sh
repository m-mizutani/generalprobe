#!/bin/bash

TEMPLATE_FILE=template.yml
OUTPUT_FILE=`mktemp`

if [ $# -le 2 ]; then
    echo "usage) $0 AwsRegion StackName CodeS3Bucket CodeS3Prefix"
    exit 1
fi

echo "Output Template: $OUTPUT_FILE"
echo ""

aws --region $1 cloudformation package --template-file $TEMPLATE_FILE \
    --output-template-file $OUTPUT_FILE --s3-bucket $3 --s3-prefix $4
aws --region $1 cloudformation deploy --template-file $OUTPUT_FILE --stack-name $2 \
    --capabilities CAPABILITY_IAM

# Resources=`aws --region $1 cloudformation describe-stack-resources --stack-name $2 | jq '.StackResources[]'`
# Lambda=`echo $Resources | jq 'select(.LogicalResourceId == "TestHandler") | .PhysicalResourceId' -r`
# DynamoDB=`echo $Resources | jq 'select(.LogicalResourceId == "ResultStore") | .PhysicalResourceId' -r`
# S3Bucket=`echo $Resources | jq 'select(.LogicalResourceId == "ResultBucket") | .PhysicalResourceId' -r`
# SNS=`echo $Resources | jq 'select(.LogicalResourceId == "Trigger") | .PhysicalResourceId' -r`

cat <<EOF > params.json
{
  "StackName": "$2",
  "Region": "$1"
}
EOF

echo ""
echo "done"
