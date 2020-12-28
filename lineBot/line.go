package main

import "github.com/line/line-bot-sdk-go/linebot"

type Line struct {
	ChannelSecret string
	ChannelToken  string
	Bot           *linebot.Client
}

func (r *Line) SendTextMessage(message string, replyToken string) error {
	return r.Reply(replyToken, linebot.NewTextMessage(message))
}

func (r *Line) SendTemplateMessage(replyToken, altText string, template linebot.Template) error {
	return r.Reply(replyToken, linebot.NewTemplateMessage(altText, template))
}

func (r *Line) Reply(replyToken string, message linebot.SendingMessage) error {
	if _, err := r.Bot.ReplyMessage(replyToken, message).Do(); err != nil {
		return err
	}
	return nil
}

func (r *Line) NewCarouselColumn(thumbnailImageURL, title, text string, actions ...linebot.TemplateAction) *linebot.CarouselColumn {
	return &linebot.CarouselColumn{
		ThumbnailImageURL: thumbnailImageURL,
		Title:             title,
		Text:              text,
		Actions:           actions,
	}
}

func (r *Line) NewCarouselTemplate(columns ...*linebot.CarouselColumn) *linebot.CarouselTemplate {
	return &linebot.CarouselTemplate{
		Columns: columns,
	}
}

func (l *Line) New(secret, token string) error {
	l.ChannelSecret = secret
	l.ChannelToken = token

	bot, err := linebot.New(
		l.ChannelSecret,
		l.ChannelToken,
	)

	if err != nil {
		return err
	}

	l.Bot = bot
	return nil
}

func (r *Line) EventRouter(eve []*linebot.Event) {
	for _, event := range eve {
		switch event.Type {
		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				r.handleText(message, event.ReplyToken, event.Source.UserID)
			}
		}

	}
}

func (r *Line) handleText(message *linebot.TextMessage, replyToken, userID string) {
	r.SendTextMessage(message.Text, replyToken)
}
