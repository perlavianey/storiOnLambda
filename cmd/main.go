package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"log"

	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/oklog/ulid/v2"
)

type Input struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	File  string `json:"file"`
}

type Transaction struct {
	Id     string
	Date   string
	Amount float64
}

func main() {
	lambda.Start(handleRequest)
}

// handleRequest is a function that handles a request to the Lambda function
func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {
	var input Input
	if err = json.Unmarshal([]byte(request.Body), &input); err != nil {
		log.Print("Error to unmarshal request body: ", err)
		response = events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    map[string]string{"Content-Type": "text/plain"},
			Body:       "Error to unmarshal request body: " + err.Error(),
		}
		return
	}

	decodedFile, err := base64.StdEncoding.DecodeString(input.File)
	if err != nil {
		log.Print("Error to decode base64 file: ", err)
		response = events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    map[string]string{"Content-Type": "text/plain"},
			Body:       "Error to decode base64 file: " + err.Error(),
		}
		return
	}

	// generate file identifier ULID
	fileIdentifier := ulid.Make()

	// read csv values using csv.Reader
	reader := csv.NewReader(bytes.NewBuffer(decodedFile))
	data, err := reader.ReadAll()
	if err != nil {
		log.Print("Error getting data from csv: ", err)
		response = events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    map[string]string{"Content-Type": "text/plain"},
			Body:       "Error getting data from csv: " + err.Error(),
		}
		return
	}

	//upload file to S3
	err = uploadFileToS3(fileIdentifier.String(), decodedFile)
	if err != nil {
		log.Print("Error to upload file ", fileIdentifier, " to S3:", err)
		response = events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    map[string]string{"Content-Type": "text/plain"},
			Body:       "Error to upload file " + fileIdentifier.String() + " to S3:" + err.Error(),
		}
		return
	}

	// convert records to array of structs
	transactionList := convertTransactions(data, fileIdentifier.String())

	//get Summary
	summary, err := getSummary(transactionList)
	if err != nil {
		log.Print("Error getting summary for customer: ", err)
		response = events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    map[string]string{"Content-Type": "text/plain"},
			Body:       "Error getting summary for customer: " + err.Error(),
		}
		return
	}

	//send email
	err = sendEmail(input.Name, input.Email, summary, decodedFile)
	if err != nil {
		log.Print("Error to send email: ", err)
		response = events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    map[string]string{"Content-Type": "text/plain"},
			Body:       "Error to send email: " + err.Error(),
		}
		return
	}

	response = events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "text/plain"},
		Body:       "Email successfully sent",
	}
	return
}
