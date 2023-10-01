package main

import (
	"errors"
	"fmt"
	"github.com/defeng-hub/Go-Email-Template"
	"github.com/go-gomail/gomail"
	"golang.org/x/crypto/ssh/terminal"
	"net/mail"
	"os"
	"strconv"
)

type example interface {
	Email() hermes.Email
	Name() string
}

func main() {
	projectName := "facec.cc"
	h := hermes.Hermes{
		DisableCSSInlining: false,                //内联css
		TextDirection:      hermes.TDLeftToRight, //从左到右
		Product: hermes.Product{
			Name: "defeng-hub",
			Link: "https://www.facec.cc/",
			Logo: "https://pic.bythedu.com/defeng_boke/%E5%BE%AE%E4%BF%A1%E5%9B%BE%E7%89%87_20220914065639_%E7%9C%8B%E5%9B%BE%E7%8E%8B_1663109992996.png",
		},
		Theme: nil,
	}
	examples := []example{
		new(welcome),
		new(reset),
		new(receipt),
		new(maintenance),
		new(inviteCode),
	}

	themes := []hermes.Theme{
		new(hermes.Default),
		new(hermes.Flat),
	}

	// Generate emails
	for _, theme := range themes {
		h.Theme = theme
		for _, e := range examples {
			generateEmails(projectName, h, e.Email(), e.Name())
		}
	}

	// 读取Email，发送Email
	sendEmails := os.Getenv("HERMES_SEND_EMAILS") == "true"
	// Send emails only when requested
	if sendEmails {
		port, _ := strconv.Atoi(os.Getenv("HERMES_SMTP_PORT"))
		password := os.Getenv("HERMES_SMTP_PASSWORD")
		SMTPUser := os.Getenv("HERMES_SMTP_USER")
		if password == "" {
			fmt.Printf("Enter SMTP password of '%s' account: ", SMTPUser)
			bytePassword, _ := terminal.ReadPassword(0)
			password = string(bytePassword)
		}
		smtpConfig := smtpAuthentication{
			Server:         os.Getenv("HERMES_SMTP_SERVER"),
			Port:           port,
			SenderEmail:    os.Getenv("HERMES_SENDER_EMAIL"),
			SenderIdentity: os.Getenv("HERMES_SENDER_IDENTITY"),
			SMTPPassword:   password,
			SMTPUser:       SMTPUser,
		}
		options := sendOptions{
			To: os.Getenv("HERMES_TO"),
		}
		for _, theme := range themes {
			h.Theme = theme
			for _, e := range examples {
				options.Subject = "Hermes | " + h.Theme.Name() + " | " + e.Name()
				fmt.Printf("Sending email '%s'...\n", options.Subject)
				htmlBytes, err := os.ReadFile(fmt.Sprintf("%v/%v.%v.html", h.Theme.Name(), h.Theme.Name(), e.Name()))
				if err != nil {
					panic(err)
				}
				txtBytes, err := os.ReadFile(fmt.Sprintf("%v/%v.%v.txt", h.Theme.Name(), h.Theme.Name(), e.Name()))
				if err != nil {
					panic(err)
				}
				err = send(smtpConfig, options, string(htmlBytes), string(txtBytes))
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func generateEmails(projectName string, h hermes.Hermes, email hermes.Email, example string) {
	dir := "email_template/" + projectName + "/"

	// Generate the HTML template and save it
	res, err := h.GenerateHTML(email)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(dir+h.Theme.Name(), 0744)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(fmt.Sprintf(dir+"%v/%v.%v.html", h.Theme.Name(), h.Theme.Name(), example), []byte(res), 0644)
	if err != nil {
		panic(err)
	}

	// Generate the plaintext template and save it
	res, err = h.GeneratePlainText(email)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(fmt.Sprintf(dir+"%v/%v.%v.txt", h.Theme.Name(), h.Theme.Name(), example), []byte(res), 0644)
	if err != nil {
		panic(err)
	}
}

// 下方都是为了发送email
type smtpAuthentication struct {
	Server         string
	Port           int
	SenderEmail    string // 发件人地址
	SenderIdentity string // 发件人姓名
	SMTPUser       string // 用户名
	SMTPPassword   string // 密码
}

// sendOptions are options for sending an email
type sendOptions struct {
	To      string
	Subject string //主题
}

// send 发送邮件
func send(smtpConfig smtpAuthentication, options sendOptions, htmlBody string, txtBody string) error {

	if smtpConfig.Server == "" {
		return errors.New("SMTP server config is empty")
	}
	if smtpConfig.Port == 0 {
		return errors.New("SMTP port config is empty")
	}

	if smtpConfig.SMTPUser == "" {
		return errors.New("SMTP user is empty")
	}

	if smtpConfig.SenderIdentity == "" {
		return errors.New("SMTP sender identity is empty")
	}

	if smtpConfig.SenderEmail == "" {
		return errors.New("SMTP sender email is empty")
	}

	if options.To == "" {
		return errors.New("no receiver emails configured")
	}

	from := mail.Address{
		Name:    smtpConfig.SenderIdentity,
		Address: smtpConfig.SenderEmail,
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from.String())
	m.SetHeader("To", options.To)
	m.SetHeader("Subject", options.Subject)

	m.SetBody("text/plain", txtBody)
	m.AddAlternative("text/html", htmlBody)

	d := gomail.NewDialer(smtpConfig.Server, smtpConfig.Port, smtpConfig.SMTPUser, smtpConfig.SMTPPassword)

	return d.DialAndSend(m)
}
