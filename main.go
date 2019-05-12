package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

type lambdaArguments struct {
	Event     events.SNSEvent
	SecretArn string
}

func getSecretValues(secretArn string, values interface{}) error {
	// sample: arn:aws:secretsmanager:ap-northeast-1:1234567890:secret:mytest
	arn := strings.Split(secretArn, ":")
	if len(arn) != 7 {
		return errors.New(fmt.Sprintf("Invalid SecretsManager ARN format: %s", secretArn))
	}
	region := arn[3]

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	mgr := secretsmanager.New(ssn)

	result, err := mgr.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretArn),
	})

	if err != nil {
		return errors.Wrap(err, "Fail to retrieve secret values")
	}

	err = json.Unmarshal([]byte(*result.SecretString), values)
	if err != nil {
		return errors.Wrap(err, "Fail to parse secret values as JSON")
	}

	return nil
}

func mainHandler(args lambdaArguments) error {
	var secrets struct {
		SlackURL string `json:"slack_url"`
	}

	if err := getSecretValues(args.SecretArn, &secrets); err != nil {
		return err
	}

	for _, record := range args.Event.Records {
		var event interface{}
		if err := json.Unmarshal([]byte(record.SNS.Message), &event); err != nil {
			return errors.Wrapf(err, "Fail to unmarshal SNS message: %s", record.SNS.Message)
		}

		logger.WithField("event", event).Info("alert")

		msg := slack.WebhookMessage{
			Username:  "Budget Monitor",
			IconEmoji: ":rotating_light:",
			Text:      "AWS Budget Alert",
			Attachments: []slack.Attachment{
				{
					Color: "#D12F2E",
					Text:  record.SNS.Message,
				},
			},
		}

		if err := slack.PostWebhook(secrets.SlackURL, &msg); err != nil {
			return err
		}
	}

	return nil
}

func handleRequest(ctx context.Context, event events.SNSEvent) error {
	logger.WithField("event", event).Info("Start")

	args := lambdaArguments{
		Event:     event,
		SecretArn: os.Getenv("SecretArn"),
	}

	if err := mainHandler(args); err != nil {
		logger.WithError(err).Error("Fail")
		return err
	}

	return nil
}

func main() {
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	lambda.Start(handleRequest)
}
