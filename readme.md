<h1>Stori Technical Challenge on AWS Lambda</h1>

This project consists on a lambda that receives a file and the name and email from the customer and builds a formatted email which is sent to the customer to let him know a brief summary of his transactions in a certain period of time.
As a feature, it's able to save the csv file in AWS S3.
<hr>
<br>
<strong>GitHub repository:</strong><br> 
https://github.com/perlavianey/storiOnLambda<br>
Release v.0.0: https://github.com/perlavianey/storiOnLambda/releases/tag/storiChallengeLambda
<br>

<strong>Built With:</strong><br>
<li>Golang as development language</li>
<li>AWS as Cloud Platform (using Lambda and S3)</li>
<li>HTML for the email template</li>
<br>

<strong>Algorithm Description:</strong><br><br>
The algorithm reads the request body, saves the specified file to S3 using [ULID](https://github.com/ulid/spec) as filename identifiers:<br><br>

![Alt text](screenshot_s3_aws.png?raw=true "AWS S3 screenshot")

<br>
Then it makes the corresponding calculations to get the total balance, transactions per month and average debit and credit amounts and then builds a formatted email with the summary, which is sent to the customer in a pretty format. The csv file is also send as an attachment to the email: <br>

![Alt text](screenshot_email.png?raw=true "Email screenshot")


<br>

<strong>Getting started</strong><br><br>
<strong>Prerrequisites:</strong><br>
To run this project you need:<br>
<ol>
<li>To have an AWS account and be able to create a new Lambda with S3 PUTObject and GETObject permissions</li>
<li>To have a S3 bucket named  <strong>transactions-stori-pv</strong> in that AWS account.</li>
<li>To have the AWS access key and secret access key to that AWS account.</li>
<li>To have a S3 bucket named  <strong>email-templates-pv</strong> in that AWS account.</li>
</ol>
<br>
<strong>Usage:</strong><br>
To run this project:<br>
<ol>
<li>Store the file <strong>template.html</strong> from cmd/email into the bucket email-templates-pv.</li>
<li>Create a new lambda called <strong>sendSummaryByEmail</strong> with the properties:<br>
Runtime: Amazon Linux 2<br>
HandlerInfo: main<br>
ArchitectureInfo: arm64</li>
<li>Set the next environment variables in the Lambda (Configuration/Environment Variables): <br>
ACCESS_KEY_ID: < PUT_HERE_YOUR_ACCESS_KEY ><br>
SECRET_ACCESS_KEY	< PUT_HERE_YOUR_SECRET_ACCESS_KER></li>
<li>Upload the file cmd/storiOnLambda.zip as the source code of the lambda (Code/Code source/Upload from .zip file)</li>
<li>Go to the option test of the Lambda and send the following event JSON changing olny your name and email address to get the email in your inbox:<br>
{
  "body": "{\"name\":\"HERE_YOUR_NAME\",\"email\":\"HERE_YOUR_EMAIL_ADDRESS\",\"file\":\"SWQsRGF0ZSxUcmFuc2FjdGlvbg0KMCwyMDIzLTA3LTE1LDYwLjUNCjEsMjAyMy0wNy0yOCwtMTAuMw0KMiwyMDIzLTA4LTAyLC0yMC40Ng0KMywyMDIzLTA4LTEzLDEw\"}"
}
</li>

</ol>

You'll get an email to the address specified in step 2, formatted with the Stori's logo, which has the Summary of the transactions indicated in the file directory/input.csv and this file as an attachment. This summary contains the total balance, the number of transactions per month and the average debit and credit amounts.