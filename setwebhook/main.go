package setwebhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/yevhenshymotiuk/drug-lord-of-armavir-bot/apigateway"
)

func handler(
	request events.APIGatewayProxyRequest,
) (apigateway.Response, error) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	url := fmt.Sprintf(
		"https://%s/%s/",
		request.Headers["Host"],
		request.RequestContext.Stage,
	)
	_, err = bot.SetWebhook(tgbotapi.NewWebhook(url))
	if err != nil {
		log.Fatal(err)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}
	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}

	var buf bytes.Buffer

	body, err := json.Marshal(
		map[string]interface{}{"message": "Webhook was successfully set!"},
	)
	if err != nil {
		return apigateway.Response{StatusCode: 404}, err
	}
	json.HTMLEscape(&buf, body)

	resp := apigateway.Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers:         map[string]string{"Content-Type": "application/json"},
	}

	return resp, nil
}

func main() {
	lambda.Start(handler)
}
