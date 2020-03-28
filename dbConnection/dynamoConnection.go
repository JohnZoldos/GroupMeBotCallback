package dbConnection

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
	"os"
)


type Item struct {
	Group_id  		string      `json:"group_id"`
	Bot_id	  		string    	`json:"bot_id"`
	Last_message_id string 		`json:"last_message_id"`
}

type ItemKey struct {
	Group_id  		string      `json:"group_id"`
}


var dynamoClient *dynamodb.DynamoDB

const tableName = "GroupMeBot"

func startSession() {
	log.Print("Dynamo session started.")
	session, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		log.Print("Error reached when starting dynamo session")
		log.Print(err)
		panic(err)
	}
	dynamoClient = dynamodb.New(session)

}

func GetInfoForGroup(groupId string) Item {
	if dynamoClient == nil {
		startSession()
	}
	log.Print("Getting group from db.")


	itemKey := ItemKey {
		Group_id: groupId,
	}
	key, err := dynamodbattribute.MarshalMap(itemKey)
	log.Println(key)
	if err != nil {
		panic(err)
	}

	input := &dynamodb.GetItemInput{
		Key:       key,
		TableName: aws.String(tableName),
	}
	// Make the DynamoDB Query API call
	result, err  := dynamoClient.GetItem(input)
	if err != nil {
		log.Print("Error reached when querying db. Exiting.")
		log.Print(err)
		os.Exit(1)
	}

	item := Item{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		panic(err)
	}

	log.Print("Got item from db.")
	log.Print(item)

	return item

}



