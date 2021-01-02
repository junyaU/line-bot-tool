package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sclevine/agouti"
)

type Line struct {
	ChannelSecret string
	ChannelToken  string
	Bot           *linebot.Client
}

type ReportData struct {
	ClassName  string
	ReportName string
	Expire     string
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

func (r *Line) ReportText(replyToken string) {
	url := "http://www.ritsumei.ac.jp/ct/"

	opts := []agouti.Option{
		agouti.Debug,
		agouti.ChromeOptions(
			"args", []string{
				// "--headless",
				"--disable-gpu",
				"--window-size=1280,800",
				"--no-sandbox",
				"--allow-insecure-localhost",
			}),
	}

	opts = append(opts, agouti.ChromeOptions(
		"binary", "/opt/headless-chromium",
	))

	driver := agouti.NewWebDriver("http://{{.Address}}", []string{"/opt/chromedriver", "--port={{.Port}}"}, opts...)

	err := driver.Start()
	if err != nil {
		log.Println(err)
	}

	defer driver.Stop()
	time.Sleep(2 * time.Second)

	page, err := driver.NewPage()

	if err != nil {
		log.Println("fail page open")
		log.Println(err)
	}

	err = page.Navigate(url)
	if err != nil {
		log.Println(err)
	}

	time.Sleep(4 * time.Second)

	page.FindByID("btnLogin").Click()
	time.Sleep(5)

	page.NextWindow()

	time.Sleep(5 * time.Second)

	//ログインに必要な情報入力
	idInputDom := page.FindByID("User_ID")
	idInputDom.Fill(os.Getenv("MANABA_ID"))
	passwordInputDom := page.FindByID("Password")
	passwordInputDom.Fill(os.Getenv("MANABA_PASSWORD"))
	submitButton := page.FindByID("Submit")
	submitButton.Click()

	time.Sleep(2 * time.Second)
	page.Navigate("https://ct.ritsumei.ac.jp/ct/home_course")

	time.Sleep(2 * time.Second)
	log.Println(page.Title())
	courses := page.AllByClass("course-cell").First("a")

	courseElements, _ := courses.Elements()

	var reportUrlsArr []string

	//各授業のレポートページのURLを配列に入れる
	for _, course := range courseElements {
		reportUrl, _ := course.GetAttribute("href")
		reportUrlsArr = append(reportUrlsArr, reportUrl+"_report")
	}

	var reportDatas []ReportData

	for _, url := range reportUrlsArr {
		page.Navigate(url)
		time.Sleep(1 * time.Second)
		dom, _ := page.HTML()
		readContents := strings.NewReader(dom)
		contentsDom, _ := goquery.NewDocumentFromReader(readContents)

		//授業名
		className := contentsDom.Find("a#coursename").Text()

		noSubmittedDoms := contentsDom.Find(".deadline").Parent().Parent()
		if noSubmittedDoms != nil {

			noSubmittedDoms.Each(func(index int, selection *goquery.Selection) {
				reportArr := ReportData{}
				reportArr.ReportName = strings.TrimSpace(selection.Find("h3.report-title").Text())
				reportArr.ClassName = className
				reportArr.Expire = strings.TrimSpace(selection.Find("td").Last().Text())
				reportDatas = append(reportDatas, reportArr)
			})
		}
	}

	if _, er := r.Bot.ReplyMessage(replyToken, linebot.NewTextMessage(reportDatas[0].ReportName)).Do(); er != nil {
		log.Println(er)
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
			switch event.Message.(type) {
			case *linebot.TextMessage:
				r.ReportText(event.ReplyToken)
			case *linebot.ImageMessage:
				r.Bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("いい写真ですね")).Do()
			case *linebot.StickerMessage:
				r.Bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("そのスタンプかわいい")).Do()
			}
		}
	}
}

func (r *Line) handleText(message *linebot.TextMessage, replyToken, userID string) {
	r.SendTextMessage(message.Text, replyToken)
}
