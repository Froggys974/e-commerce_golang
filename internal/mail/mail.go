package mail

import (
	"errors"
	"fmt"
	"log"
	"mime"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type Sender struct {
	address string // host:port serveur SMTP
	from    string
}

func NewSender(host, port, from string) *Sender {
	return &Sender{address: net.JoinHostPort(host, port), from: from}
}

const sendTimeout = 10 * time.Second

func (sender *Sender) SendCode(to, subject, code string) error {
	message, err := buildMessage(sender.from, to, subject, code)
	if err != nil {
		return err
	}
	if err := sender.send(to, message); err != nil {
		return fmt.Errorf("envoi du mail à %s : %w", to, err)
	}
	return nil
}

// comme smtp SendMail mais avec un delai maximal
func (sender *Sender) send(to string, message []byte) error {
	conn, err := net.DialTimeout("tcp", sender.address, sendTimeout)
	if err != nil {
		return err
	}
	if err := conn.SetDeadline(time.Now().Add(sendTimeout)); err != nil {
		conn.Close()
		return err
	}

	host, _, err := net.SplitHostPort(sender.address)
	if err != nil {
		conn.Close()
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return err
	}
	defer client.Close()

	if err := client.Mail(sender.from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(message); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func buildMessage(from, to, subject, code string) ([]byte, error) {
	if strings.ContainsAny(from+to, "\r\n") {
		return nil, errors.New("adresse email invalide")
	}

	headers := []string{
		"From: " + from,
		"To: " + to,
		"Subject: " + mime.QEncoding.Encode("utf-8", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
	}
	body := fmt.Sprintf("Bonjour,\r\n\r\nVotre code : %s\r\n\r\nIl est valable 15 minutes.\r\n", code)

	return []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + body), nil
}

type LogSender struct{}

func (LogSender) SendCode(to, subject, code string) error {
	log.Printf("[MAIL] %s pour %s : %s", subject, to, code)
	return nil
}

type CodeSender interface {
	SendCode(to, subject, code string) error
}

func FromEnv() CodeSender {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		log.Println("SMTP_HOST vide : codes affichés dans les logs")
		return LogSender{}
	}

	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "1025"
	}
	from := os.Getenv("MAIL_FROM")
	if from == "" {
		from = "no-reply@ecommerce.local"
	}
	log.Printf("Mails envoyés via SMTP %s:%s", host, port)
	return NewSender(host, port, from)
}
