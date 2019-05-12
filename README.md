# aws-budget-alert-slack

Notification sendor of AWS Budget Alert.

## Prerequisite

- go >= 1.12.4
- awscli >= 1.16.140
- GNU make >= 3.81

## Deploy

### Create a Slack Incoming Webhook URL

See https://api.slack.com/incoming-webhooks

### Create a secret of AWS SecretsManager

See https://docs.aws.amazon.com/secretsmanager/latest/userguide/manage_create-basic-secret.html

Create your secret as `Other type of secret`. The secret must contain `slack_url` with Slack Incomding Webhook URL.

### Create a config file and deploy.

Prepare a config file like following and save it as `config.json`.

```json
{
    "StackName": "budget-alert",
    "Region": "ap-northeast-1",
    "CodeS3Bucket": "mizutani-security-log.mgmt",
    "CodeS3Prefix": "functions",
    "SecretArn": "arn:aws:secretsmanager:ap-northeast-1:1234567890:secret:aws-budget-alert-XXXX"
}
```

Then, run following command.

```bash
$ make deploy
```