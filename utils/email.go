package utils

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"time"

	"github.com/anishmgoyal/calagora-admin/models"
	"github.com/microcosm-cc/bluemonday"
)

const maxPartSize = 1024 * 1024 * 25

// RawEmail contains fields for an email as it is being
// parsed
type RawEmail struct {
	Email       models.Email
	Attachments []RawAttachment
}

// RawAttachment contains data that can be used to construct an attachment.
// These have yet to be saved to disk
type RawAttachment struct {
	ContentType  string
	FileName     string
	FileContents []byte
}

type headerInterface interface {
	Get(key string) string
}

var policy *bluemonday.Policy

func initEmail() {
	policy = bluemonday.UGCPolicy()
}

// ParseEmail attempts to parse an email
func ParseEmail(contents string) *models.Email {
	email := models.Email{}
	reader := strings.NewReader(contents)
	message, err := mail.ReadMessage(reader)
	header := message.Header
	received, err := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700",
		header.Get("Date"))
	if err != nil {
		received = time.Now()
	}
	email.Received = received
	email.From = header.Get("From")
	email.Subject = header.Get("Subject")

	email.IsSpam = strings.Compare(
		header.Get("X-SES-Spam-Verdict"), "PASS") != 0
	email.IsVirus = strings.Compare(
		header.Get("X-SES-Virus-Verdict"), "PASS") != 0

	email.To = strings.Split(header.Get("To"), ",")
	email.CC = strings.Split(header.Get("Cc"), ",")
	email.BCC = strings.Split(header.Get("Bcc"), ",")

	for _, slice := range [][]string{email.To, email.CC, email.BCC} {
		for i, addr := range slice {
			slice[i] = strings.TrimSpace(addr)
			idxStart := strings.Index(addr, "<")
			idxEnd := strings.Index(addr, ">")
			if idxEnd > idxStart+5 {
				slice[i] = addr[idxStart+1 : idxEnd]
			}
		}
	}
	/*for i, addr := range email.To {
		email.To[i] = strings.TrimSpace(addr)
		idxStart := strings.Index(addr, "<")
		idxEnd := strings.Index(addr, ">")
		if idxEnd > idxStart+5 {
			email.To[i] = addr[idxStart+1 : idxEnd]
		}
	}
	for i, addr := range email.CC {
		email.CC[i] = strings.TrimSpace(addr)
		idxStart := strings.Index(addr, "<")
		idxEnd := strings.Index(addr, ">")
		if idxEnd > idxStart+5 {
			email.CC[i] = addr[idxStart+1 : idxEnd]
		}
	}*/

	mediaType, params, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil {

		// We don't parse quoted-printable later, so we need to
		// parse it here.
		reader := message.Body
		switch strings.ToLower(header.Get("Content-Transfer-Encoding")) {
		case "quoted-printable":
			reader = quotedprintable.NewReader(reader)
		}

		// We may have only been given a formatted component
		if strings.Compare(header.Get("Content-Type"), "text/html") == 0 {
			parseFormattedText(&email, reader, header)
		} else {
			parsePlainText(&email, reader, header)
		}

	} else {

		if strings.HasPrefix(mediaType, "multipart/") {
			parseMultipart(&email, message.Body, params)
		} else if strings.HasPrefix(mediaType, "text/") {

			if strings.Compare(mediaType, "text/html") == 0 {
				parsePlainText(&email, message.Body, header)
			} else if len(email.PlainText) == 0 || strings.Compare(mediaType,
				"text/plain") == 0 {

				parsePlainText(&email, message.Body, header)
			}

		}
	}

	return &email
}

func parseMultipart(email *models.Email, body io.Reader,
	params map[string]string) {

	mr := multipart.NewReader(body, params["boundary"])
	for {
		part, err := mr.NextPart()
		if err != nil {
			break
		}

		isEmailBody := len(part.FileName()) == 0
		header := part.Header
		mediaType, partParams, err := mime.ParseMediaType(
			header.Get("Content-Type"))

		if err != nil {
			parsePlainText(email, part, header)
		} else {

			if strings.HasPrefix(mediaType, "multipart/") {
				parseMultipart(email, part, partParams)

			} else if isEmailBody && strings.Compare(mediaType, "text/plain") == 0 {
				parsePlainText(email, part, header)

			} else if isEmailBody && strings.Compare(mediaType, "text/html") == 0 {
				parseFormattedText(email, part, header)

			} else {

				bytes := readBytesFromReader(part, header)
				if err == nil {
					attachment := models.Attachment{
						ContentType: mediaType,
						FileName:    part.FileName(),
						RawData:     bytes,
					}
					email.Attachments = append(email.Attachments, attachment)
				}

			}
		}
	}
}

func parsePlainText(email *models.Email, body io.Reader,
	header headerInterface) {

	bytes := readBytesFromReader(body, header)
	email.PlainText = strings.TrimSpace(string(bytes))
}

func parseFormattedText(email *models.Email, body io.Reader,
	header headerInterface) {

	bytes := readBytesFromReader(body, header)
	text := strings.TrimSpace(string(bytes))

	// A provided policy for sanitizing input
	email.FormattedText = text //policy.Sanitize(text)
}

func readBytesFromReader(reader io.Reader, header headerInterface) []byte {

	switch strings.ToLower(header.Get("Content-Transfer-Encoding")) {
	case "base64":
		reader = base64.NewDecoder(base64.StdEncoding, reader)
	}

	var buff bytes.Buffer
	var slice = make([]byte, 1024)
	total := 0

	for total < maxPartSize {
		n, err := reader.Read(slice)
		if err == io.EOF {
			break
		}
		if err != nil {
			return []byte{}
		}
		total += n
		buff.Write(slice[0:n])
	}
	return buff.Bytes()
}
