package main

import (
	"GroupMeBotCallback/dbConnection"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/subosito/gotenv"
)

const URL_BASE = "https://api.groupme.com/v3"
const BOT_NAME = "MemsBot"
const MAX_MESSAGE_LEN = 1000

type Body struct {
	GroupId    string `json:"group_id"`
	SenderId   string `json:"sender_id"`
	SenderType string `json:"sender_type"`
	Text       string `json:"text"`
}

type Attachment struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}

type Message struct {
	Name        string       `json:"name"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`

	numMembersAtTime int
}

type Messages struct {
	Messages []*Message `json:"messages"`
}

type MessagesResponse struct {
	MessagesMap Messages `json:"response"`
}

func getMessageBatch(groupId string, accessToken string, before_id string, numMessages int) []*Message {
	url := fmt.Sprintf("%s/groups/%s/messages?token=%s&limit=%d&before_id=%s", URL_BASE, groupId, accessToken, numMessages, before_id)
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		log.Print("Fatal error reached when getting message batch.")
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	log.Println("Body:")
	log.Println(string(body))
	if err != nil {
		log.Print("Fatal error reached when reading message batch.")
		log.Fatalln(err)
	}
	messageResponse := MessagesResponse{}
	err = json.Unmarshal(body, &messageResponse)
	if err != nil {
		log.Print("Fatal error reached when unmarshalling message batch.")
		panic(err)
	}
	messagesBatch := messageResponse.MessagesMap.Messages
	log.Println("Batch:")
	log.Println(messagesBatch)

	return messagesBatch

}

func getGroupAndNumMessages(req events.APIGatewayProxyRequest) (string, int) {
	body := Body{}
	err := json.Unmarshal([]byte(req.Body), &body)
	if err != nil {
		log.Print("Fatal error reached when unmarshaling body.")
		panic(err)
	}
	log.Println(body.GroupId)
	log.Println(body.SenderId)
	log.Println(body.SenderType)
	if body.SenderType == "bot" {
		os.Exit(0)
	}
	text := strings.ToLower(body.Text)
	numMessages := -1
	re := regexp.MustCompile(`^@memsbot \d$`)
	matches := re.FindStringSubmatch(text)
	if matches != nil {
		match := matches[0]
		number := match[len(match)-1:]
		numMessages, _ = strconv.Atoi(number)
	}
	return body.GroupId, numMessages
}

func getAccessToken() string {
	gotenv.Load()
	accessToken := os.Getenv("ACCESS_TOKEN")
	return accessToken
}

func postMessages(messages []*Message, botId string) {
	header := "Last Mem's Context:"
	context := ""
	prevContextVal := context
	for _, message := range messages {
		if strings.HasPrefix(message.Text, header) {
			continue
		}
		for _, attachement := range message.Attachments {
			context = " " + attachement.Url + context
		}
		sender := message.Name
		text := message.Text
		context = fmt.Sprintf("\n- %s: %s", sender, text) + context
		if len(context) > MAX_MESSAGE_LEN {
			context = prevContextVal
			break
		}
		prevContextVal = context
	}

	context = header + context

	url := fmt.Sprintf("%s/bots/post", URL_BASE)
	params := map[string]interface{}{
		"bot_id": botId,
		"text":   context,
	}
	log.Println("params:")
	log.Println(params)
	bytesRepresentation, err := json.Marshal(params)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Print("Fatal error reached when posting message.")
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	log.Print("Post message api request completed.")

}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	gotenv.Load()
	log.Println(req)
	groupId, numMessages := getGroupAndNumMessages(req)
	if numMessages < 1 {
		os.Exit(0)
	}
	item := dbConnection.GetInfoForGroup(groupId)

	messages := getMessageBatch(groupId, getAccessToken(), item.Last_message_id, numMessages)
	log.Println(messages)

	postMessages(messages, item.Bot_id)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(req.Body),
	}, nil
}

func main() {
	lambda.Start(handler)

}
