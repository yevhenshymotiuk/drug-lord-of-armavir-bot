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
)

type Response events.APIGatewayProxyResponse

var okResp = Response{
	StatusCode:      200,
	IsBase64Encoded: false,
	Body:            "Ok",
}

func getObjectFromS3Bucket(bucketName string, objectName string) *s3.GetObjectOutput {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("eu-north-1"),
	})

	client := s3.New(sess)

	resp, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	})

	if err != nil {
		log.Fatalf("Unable to get file %q, %v", objectName, err)
	}

	return resp
}

func s3ObjectToAudioFile(bucketName string, objectName string) tgbotapi.FileReader {
	resp := getObjectFromS3Bucket(
		bucketName, objectName)
	audioFile := tgbotapi.FileReader{
		Name:   objectName,
		Reader: io.Reader(resp.Body),
		Size:   -1,
	}

	return audioFile
}

func Handler(request events.APIGatewayProxyRequest) (Response, error) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	update := tgbotapi.Update{}

	bodyUnmarshalErr := json.Unmarshal([]byte(request.Body), &update)
	if bodyUnmarshalErr != nil {
		log.Panic(bodyUnmarshalErr)
	}

	if update.Message == nil { // ignore any non-Message Updates
		return okResp, nil
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
		log.Panic(err)
	}

	return okResp, nil
}

func main() {
	lambda.Start(Handler)
}
