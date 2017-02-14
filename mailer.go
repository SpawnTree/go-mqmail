//If you are using this package to send a simple email create a NewRequest and use it in the SendEmail API. Otherwise create a NewRequest without setting the body and setting the templateData and templateFileName and then call ParseTemplate
package main

import (
	"bytes"

	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"mime"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
)

//Request is an struct that contains the info for sending an email
type Request struct {
	From             string
	ReplyTo          string
	To               []string
	CC               []string
	Bcc              []string
	Subject          string
	Body             string
	HTML             bool
	TemplateFileName string
	TemplateData     interface{}
	Attachments      map[string]*Attachment
}

//Attachment is the struct for capturing attachments
type Attachment struct {
	Filename string
	Data     []byte
}

var auth smtp.Auth

const (
	MailHost string = "smtp.mailgun.org"
	UserName string = "postmaster@sandbox7c55d36539c9469893b865a0faf7ad43.mailgun.org"
	Password string = "7f12d9c566f12f7a38fb5c5d1a7f76de"
	MailPort string = "25"
)

//NewRequestWithReplyTo creates a new request object with ReplyTo. If you are sending a simple email create a NewRequest and set the html as false. Otherwise if you want to send an html email pass the html value as true.
func NewRequestWithReplyTo(to, cc, bcc []string, from, replyTo, subject, body string, html bool) *Request {
	r := &Request{
		To:      to,
		CC:      cc,
		Bcc:     bcc,
		Subject: subject,
		From:    from,
		ReplyTo: replyTo,
		Body:    body,
		HTML:    html,
	}
	r.Attachments = make(map[string]*Attachment)
	return r
}

//NewRequest creates a new request object. If you are sending a simple email create a NewRequest and set the html as false. Otherwise if you want to send an html email pass the html value as true.
func NewRequest(to, cc, bcc []string, from, subject, body string, html bool) *Request {
	return NewRequestWithReplyTo(to, cc, bcc, from, from, subject, body, html)
}

//Init initializes the mail system
func Init() {
	auth = smtp.PlainAuth("", UserName, Password, MailHost)
}

//Attach attachs a file to the request
func (r *Request) Attach(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	_, filename := filepath.Split(file)

	r.Attachments[filename] = &Attachment{
		Filename: filename,
		Data:     data,
	}
	return nil
}

func getMessageWithAttachment(mimeStr string, r *Request) ([]string, []byte) {
	buf := bytes.NewBuffer(nil)
	from := "From: " + r.From + "\n"
	buf.WriteString(from)
	var to, cc string
	var recipient []string

	if r.To != nil {
		to = "To: " + strings.Join(r.To, ", ") + "\n"
		recipient = append(recipient, r.To...)
		buf.WriteString(to)
	}
	if r.CC != nil {
		cc = "Cc: " + strings.Join(r.CC, ", ") + "\n"
		recipient = append(recipient, r.CC...)
		buf.WriteString(cc)
	}
	subject := "Subject: " + r.Subject + "\n"
	buf.WriteString(subject)
	buf.WriteString("MIME-Version: 1.0\r\n")
	boundary := "f46d043c813270fc6b04c2d223da"
	buf.WriteString("Content-Type: multipart/mixed; boundary=" + boundary + "\r\n")
	buf.WriteString("--" + boundary + "\r\n")

	buf.WriteString(mimeStr)
	buf.WriteString(r.Body)
	buf.WriteString("\r\n")
	for _, attachment := range r.Attachments {
		buf.WriteString("\r\n\r\n--" + boundary + "\r\n")

		ext := filepath.Ext(attachment.Filename)
		mimetype := mime.TypeByExtension(ext)
		if mimetype != "" {
			mime := fmt.Sprintf("Content-Type: %s\r\n", mimetype)
			buf.WriteString(mime)
		} else {
			buf.WriteString("Content-Type: application/octet-stream\r\n")
		}
		// buf.WriteString("Content-Type: multipart/mixed; boundary=" + boundary)
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString("Content-Disposition: attachment; filename=\"" + attachment.Filename + "\"\r\n\r\n")

		b := make([]byte, base64.StdEncoding.EncodedLen(len(attachment.Data)))
		base64.StdEncoding.Encode(b, attachment.Data)
		// write base64 content in lines of up to 76 chars
		for i, l := 0, len(b); i < l; i++ {
			buf.WriteByte(b[i])
			if (i+1)%76 == 0 {
				buf.WriteString("\r\n")
			}
		}
		buf.WriteString("\r\n--" + boundary)
	}
	buf.WriteString("--")
	return recipient, buf.Bytes()
}

//SendEmail sends the email based on the request
func (r *Request) SendEmail() (bool, error) {
	if os.Getenv("GO_ENV") == "test" {
		return true, nil
	}
	//TODO : html templates
	mimeStr := "Content-Type: text/plain; charset=\"UTF-8\";\r\n\r\n"
	if r.HTML {
		mimeStr = "Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	}

	addr := MailHost + ":" + MailPort
	var msg []byte
	var recipient []string
	if len(r.Attachments) > 0 {
		recipient, msg = getMessageWithAttachment(mimeStr, r)
	} else {
		from := "From: " + r.From + "\r\nReply-To: " + r.ReplyTo + "\r\n"
		var to, cc string
		if r.To != nil {
			to = "To: " + strings.Join(r.To, ", ") + "\r\n"
			recipient = append(recipient, r.To...)
		}
		if r.CC != nil {
			cc = "Cc: " + strings.Join(r.CC, ", ") + "\r\n"
			recipient = append(recipient, r.CC...)
		}
		subject := "Subject: " + r.Subject + "\n"
		msg = []byte(from + to + cc + subject + mimeStr + "\n" + r.Body)
	}

	if r.Bcc != nil {
		recipient = append(recipient, r.Bcc...)
	}
	if err := smtp.SendMail(addr, auth, UserName, recipient, msg); err != nil {
		return false, err
	}
	return true, nil
}

//ParseTemplate the template files in the Request and sets the data from Request.templateData
//returns the Request with the body set
func (r *Request) ParseTemplate(templateFileName string, data interface{}) error {
	if os.Getenv("GO_ENV") == "devtest" {
		cwd, _ := os.Getwd()
		var templatePath = strings.Split(cwd, "/gtcode")[0] + "/gtcode/src/mail/"
		templateFileName = templatePath + templateFileName
	}
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}
	r.Body = buf.String()
	return nil
}
