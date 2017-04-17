package constants

import (
	"os"
	"strconv"
	"strings"
)

// PortNum is the port on which the server shall run
var PortNum = 2767

// SSLPortNum is the port on which SSL should be handled if enabled
var SSLPortNum = 2773

// SSLEnable determines whether or not the server will use SSL
var SSLEnable = false

// SSLKeyFile is the path to a key file if SSL is used
var SSLKeyFile = ""

// SSLCertificate is the path to a certificate if SSL is used
var SSLCertificate = ""

// DatabaseUsername is the username for the database connection
var DatabaseUsername = "calagorauser"

// DatabasePassword is the password for the database connection
var DatabasePassword = "calagorapassword"

// DatabaseHost is the hostname and port number for the database connection
var DatabaseHost = "localhost:5432"

// DatabaseExtraArgs is a query string with any connection parameters
var DatabaseExtraArgs = "?sslmode=disable"

// DoSendEmails decides if emails are sent out by the application
var DoSendEmails = false

// S3URLStub is the URL path to the AWS S3 bucket
var S3URLStub = "http://calagora-email.s3.amazonaws.com/"

// S3Bucket is the name of the AWS S3 bucket for uploads
var S3Bucket = "calagora-email"

// S3RegionString is the region in which we are using S3
var S3RegionString = "us-east-1"

// SMTPHostname is the server which handles sending emails
var SMTPHostname = ""

// SMTPPort is the port for the server sending emails
var SMTPPort = "587"

// SMTPAuthUser is the username for smtp
var SMTPAuthUser = ""

// SMTPAuthPassword is the password for smtp
var SMTPAuthPassword = ""

// LoadEnvironmentSettings loads settings from environment variables
func LoadEnvironmentSettings() {
	loadIntSetting(&PortNum, "CALAGORA_PORT_NUM")
	loadIntSetting(&SSLPortNum, "CALAGORA_SSL_PORT")

	loadBooleanSetting(&SSLEnable, "CALAGORA_SSL_ENABLE")
	loadStringSetting(&SSLKeyFile, "CALAGORA_SSL_KEYFILE")
	loadStringSetting(&SSLCertificate, "CALAGORA_SSL_CERTIFICATE")

	loadStringSetting(&DatabaseUsername, "CALAGORA_DB_UNAME")
	loadStringSetting(&DatabasePassword, "CALAGORA_DB_PWORD")
	loadStringSetting(&DatabaseHost, "CALAGORA_DB_HOST")
	loadStringSetting(&DatabaseExtraArgs, "CALAGORA_DB_ARGS")

	loadBooleanSetting(&DoSendEmails, "CALAGORA_SEND_EMAILS")

	loadStringSetting(&S3URLStub, "CALAGORA_S3_URL")
	loadStringSetting(&S3Bucket, "CALAGORA_S3_BUCKET")
	loadStringSetting(&S3RegionString, "CALAGORA_S3_REGION")

	loadStringSetting(&SMTPHostname, "CALAGORA_SMTP_HOST")
	loadStringSetting(&SMTPPort, "CALAGORA_SMTP_PORT")
	loadStringSetting(&SMTPAuthUser, "CALAGORA_SMTP_USER")
	loadStringSetting(&SMTPAuthPassword, "CALAGORA_SMTP_PASS")
}

func loadBooleanSetting(setting *bool, envKey string) {
	envVal := os.Getenv(envKey)
	if strings.Compare(envVal, "true") == 0 {
		*setting = true
	} else if strings.Compare(envVal, "false") == 0 {
		*setting = false
	}
}

func loadStringSetting(setting *string, envKey string) {
	envVal := os.Getenv(envKey)
	if len(envVal) > 0 {
		*setting = envVal
	}
}

func loadIntSetting(setting *int, envKey string) {
	envVal := os.Getenv(envKey)
	if len(envVal) > 0 {
		val, err := strconv.Atoi(envVal)
		if err == nil {
			*setting = val
		}
	}
}
