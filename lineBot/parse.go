package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/line/line-bot-sdk-go/linebot"
)

func ParseRequest(channelSecret string, r events.APIGatewayProxyRequest) ([]*linebot.Event, error) {

	if !validateSignature(channelSecret, r.Headers["X-Line-Signature"], []byte(r.Body)) {
		return nil, linebot.ErrInvalidSignature
	}

	request := &struct {
		Events []*linebot.Event `json:"events"`
	}{}

	if err := json.Unmarshal([]byte(r.Body), request); err != nil {
		return nil, err
	}
	return request.Events, nil
}

func validateSignature(channelSecret, signature string, body []byte) bool {
	decoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}

	hash := hmac.New(sha256.New, []byte(channelSecret))
	hash.Write(body)
	return hmac.Equal(decoded, hash.Sum(nil))
}
