package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Mail struct {
	Sender  string
	To      []string
	Subject string
	Body    string
}

// convertTransactions is a function that converts a csv file to an array of database.Transaction
func convertTransactions(data [][]string, ulid string) []Transaction {
	var transactions []Transaction
	for i, line := range data {
		if i > 0 { // omit header line
			var rec Transaction
			for j, field := range line {
				switch j {
				case 0:
					rec.Id = field
				case 1:
					dateTime, _ := time.Parse("2006-01-02", field)
					rec.Date = dateTime.Format("2006-01-02")
				case 2:
					rec.Amount, _ = strconv.ParseFloat(field, 64)
				}
			}
			transactions = append(transactions, rec)
		}

	}
	return transactions
}

// getUTCTimeFormat returns the input date in UTC time format
func getUTCTimeFormat(date time.Time) string {
	layout := "2006-01-02T15:04:05.000Z07:00"

	formattedDate := date.Format(layout)
	return formattedDate
}

// calculateTotalBalance is a function that calculates the total balance of a list of transactions
func calculateTotalBalance(transactionList []Transaction) (total float64, e error) {
	for _, transaction := range transactionList {
		total += transaction.Amount
	}
	return
}

// calculateTransactionsPerMonth is a function that divides a list of transactions by month and returns them into a map of transactions grouped by month
func calculateTransactionsPerMonth(transactionList []Transaction) (transactionsPerMonth map[string]int, e error) {
	transactionsPerMonth = make(map[string]int)

	for _, transaction := range transactionList {
		date, _ := time.Parse("2006-01-02", transaction.Date)
		m := time.Month(date.Month())
		transactionsPerMonth[m.String()]++
	}
	return
}

// calculateAverageDebit is a function that calculates the average debit amount of a list of transactions
func calculateAverageDebit(transactionList []Transaction) (average float64, e error) {
	var counter int
	for _, transaction := range transactionList {
		if transaction.Amount < 0 {
			counter++
			average += transaction.Amount
		}
	}
	average = average / float64(counter)
	return
}

// calculateAverageCredit is a function that calculates the average credit amount of a list of transactions
func calculateAverageCredit(transactionList []Transaction) (average float64, e error) {
	var counter int
	for _, transaction := range transactionList {
		if transaction.Amount > 0 {
			counter++
			average += transaction.Amount
		}
	}
	average = average / float64(counter)
	return
}

// getSummary is a function that calculates the summary of a list of transactions and returns them into a slice of strings, ready to print on the email
func getSummary(transactionList []Transaction) ([]string, error) {
	var summary []string
	//calculate summary
	//calculate totalBalance
	totalBalance, e := calculateTotalBalance(transactionList)
	if e != nil {
		return nil, e
	}
	summary = append(summary, fmt.Sprintf("Total balance is: %.2f", totalBalance))

	//calculate transactions per month
	transactionsPerMonth, e := calculateTransactionsPerMonth(transactionList)
	if e != nil {
		return nil, e
	}

	for key, value := range transactionsPerMonth {
		summary = append(summary, fmt.Sprintf("Number of transactions in %v: %v\n", key, value))
	}

	//calculate average debit
	averageDebit, e := calculateAverageDebit(transactionList)
	if e != nil {
		return nil, e
	}
	summary = append(summary, fmt.Sprintf("Average debit amount: %.2f", averageDebit))

	//calculate average credit
	averageCredit, e := calculateAverageCredit(transactionList)
	if e != nil {
		return nil, e
	}
	summary = append(summary, fmt.Sprintf("Average credit amount: %.2f", averageCredit))

	return summary, nil
}

// sendEmail is a function that sends an email with the summary of a list of transactions
func sendEmail(name string, email string, summary []string, fileByte []byte) error {
	sender := "storitests@gmail.com"
	password := "eixy zwpd olde vsrn"

	to := []string{
		email,
	}

	subject := "Summary from Stori"

	request := Mail{
		Sender:  sender,
		To:      to,
		Subject: subject,
	}

	host := "smtp.gmail.com"
	addr := "smtp.gmail.com:587"

	data := buildMail(name, summary, request, fileByte)
	auth := smtp.PlainAuth("", sender, password, host)
	err := smtp.SendMail(addr, auth, sender, to, data)

	if err != nil {
		return err
	}

	fmt.Println("Email sent successfully")
	return nil
}

// buildMail is a function that builds the email body from a template and returns it as a byte array
func buildMail(name string, summary []string, mail Mail, fileByte []byte) []byte {

	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("From: %s\r\n", mail.Sender))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(mail.To, ";")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", mail.Subject))

	boundary := "my-boundary-779"
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n",
		boundary))

	buf.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))

	buf.WriteString("MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n")
	err := getFileFromS3()
	if err != nil {
		log.Print(err)
	}
	t, err := template.ParseFiles("/tmp/template.html")
	if err != nil {
		log.Print(err)
	}
	t.Execute(&buf, struct {
		Name    string
		Message []string
	}{
		Name:    name,
		Message: summary,
	})

	buf.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	buf.WriteString("Content-Transfer-Encoding: base64\r\n")
	buf.WriteString("Content-Disposition: attachment; filename=transactions.csv\r\n")
	buf.WriteString("Content-ID: <transactions.csv>\r\n\r\n")

	b := make([]byte, base64.StdEncoding.EncodedLen(len(fileByte)))
	base64.StdEncoding.Encode(b, fileByte)
	buf.Write(b)
	buf.WriteString(fmt.Sprintf("\r\n--%s", boundary))

	buf.WriteString("--")

	return buf.Bytes()
}

// getFileFromS3 is a function that gets the template.html file from S3 and saves it locally in /tmp/template.html
func getFileFromS3() (err error) {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)
	downloader := s3manager.NewDownloader(sess)

	file, err := os.Create("/tmp/template.html")
	if err != nil {
		log.Fatal("Unable to create template in os:", err)
		return err
	}
	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String("email-templates-pv"),
			Key:    aws.String("template.html"),
		})
	if err != nil {
		log.Fatal("Unable to download template:", err)
		return err
	}

	return
}

// uploadFileToS3 is a function that uploads a file to an S3 bucket
func uploadFileToS3(fileIdentifier string, fileByte []byte) error {
	KeyId := os.Getenv("ACCESS_KEY_ID")
	SecretKey := os.Getenv("SECRET_ACCESS_KEY")
	s3Config := &aws.Config{
		Region:      aws.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials(KeyId, SecretKey, ""),
	}
	s3Session := session.New(s3Config)

	uploader := s3manager.NewUploader(s3Session)
	input := &s3manager.UploadInput{
		Bucket:      aws.String("transactions-stori-pv"),                        // bucket's name
		Key:         aws.String(fileIdentifier + "/" + fileIdentifier + ".csv"), // file key
		Body:        bytes.NewReader(fileByte),                                  // body of the file
		ContentType: aws.String("text/csv"),                                     // content type
	}
	_, err := uploader.UploadWithContext(context.Background(), input)
	if err != nil {
		return err
	}
	return nil
}
