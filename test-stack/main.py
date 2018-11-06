import json
import boto3
import logging
import os
import datetime


logger = logging.getLogger()
logger.setLevel(level=logging.INFO)


def iter_sns_msg(records):
    for record in records.get('Records', []):
        try:
            sns_msg = record.get('Sns', {}).get('Message')
            yield json.loads(sns_msg)
        except Exception as e:
            logger.error(e)


def handler(records, context):
    dybamo_client = boto3.client('dynamodb')
    s3_client = boto3.client('s3')

    for msg in iter_sns_msg(records):
        logger.info(json.dumps(msg, indent=2))

        dynamo_res = dybamo_client.put_item(
            TableName=os.environ["TABLE_NAME"],
            Item={
                'result_id': {'S': msg['id']},
                'report': {'B': json.dumps(msg)},
            })

        logger.info('DynamoDB: {}'.format(dynamo_res))

        s3_res = s3_client.put_object(
            Bucket=os.environ['BUCKET_NAME'],
            Key='{}/data.json'.format(msg['id']),
            Body=json.dumps(msg).encode('utf8')
        )
        logger.info('S3: {}'.format(s3_res))

    return {'message': 'ok'}
