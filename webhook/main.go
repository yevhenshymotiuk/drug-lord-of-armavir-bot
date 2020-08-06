package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/yevhenshymotiuk/telegram-lambda-helpers/apigateway"
)

func getObjectFromS3Bucket(
	bucketName string,
	objectName string,
) *s3.GetObjectOutput {
	sess, _ := session.NewSession(&aws.Config{Region: aws.String("eu-north-1")})

	client := s3.New(sess)

	resp, err := client.GetObject(
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectName),
		},
	)

	if err != nil {
		log.Fatalf("Unable to get file %q, %v", objectName, err)
	}

	return resp
}

func s3ObjectToAudioFile(
	bucketName string,
	objectName string,
) tgbotapi.FileReader {
	resp := getObjectFromS3Bucket(bucketName, objectName)
	audioFile := tgbotapi.FileReader{
		Name:   objectName,
		Reader: io.Reader(resp.Body),
		Size:   -1,
	}

	return audioFile
}

func handler(
	request events.APIGatewayProxyRequest,
) (apigateway.Response, error) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		return apigateway.Response404, err
	}

	update := tgbotapi.Update{}

	err = json.Unmarshal([]byte(request.Body), &update)
	if err != nil {
		return apigateway.Response404, err
	}

	if update.Message == nil { // ignore any non-Message Updates
		return apigateway.Response200, nil
	}

	var msg tgbotapi.VoiceConfig
	assetsBucket := os.Getenv("ASSETS_BUCKET")

	if update.Message.Command() == "start" {
		// Send greeting voice message
		audioFile := s3ObjectToAudioFile(assetsBucket, "greeting.ogg")
		msg = tgbotapi.NewVoiceUpload(update.Message.Chat.ID, audioFile)
		msg.Duration = 1
	} else {
		// Send "A, da" voice message
		s := rand.NewSource(time.Now().UnixNano())
		r := rand.New(s)

		s3ObjectName := fmt.Sprintf("a-da-%d.ogg", r.Intn(3))
		audioFile := s3ObjectToAudioFile(assetsBucket, s3ObjectName)
		msg = tgbotapi.NewVoiceUpload(update.Message.Chat.ID, audioFile)
		msg.Duration = 1
		msg.ReplyToMessageID = update.Message.MessageID
	}

	_, err = bot.Send(msg)
	if err != nil {
		return apigateway.Response404, err
	}

	return apigateway.Response200, nil
}

func main() {
	lambda.Start(handler)
}
