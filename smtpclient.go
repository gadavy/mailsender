package mailsender

import (
	"bytes"
	"crypto/tls"
	"errors"
	"net"
	"net/mail"
	"net/smtp"

	jordanEmail "github.com/jordan-wright/email"
)

type SMTPClientBuilder interface {
	Host(host string) SMTPClientBuilder
	Login(login string) SMTPClientBuilder
	Password(password string) SMTPClientBuilder
	TLS(isTLS bool) SMTPClientBuilder
	SSL(isSSL bool) SMTPClientBuilder
	Build() (*smtpClient, error)
}

func NewSMTPClient() SMTPClientBuilder {
	return &smtpClientBuilder{
		client: &smtpClient{
			client: jordanEmail.NewEmail(),
		},
	}
}

type SMTPClient interface {
	Send(*email) error
}

type smtpClientBuilder struct {
	client *smtpClient
}

func (scb *smtpClientBuilder) Host(host string) SMTPClientBuilder {
	scb.client.Host = host
	return scb
}

func (scb *smtpClientBuilder) Login(login string) SMTPClientBuilder {
	scb.client.Login = login
	return scb
}

func (scb *smtpClientBuilder) Password(password string) SMTPClientBuilder {
	scb.client.Password = password
	return scb
}

func (scb *smtpClientBuilder) TLS(isTLS bool) SMTPClientBuilder {
	scb.client.TLS = isTLS
	return scb
}

func (scb *smtpClientBuilder) SSL(isSSL bool) SMTPClientBuilder {
	scb.client.SSL = isSSL
	return scb
}

func (scb *smtpClientBuilder) Build() (*smtpClient, error) {
	_, _, err := net.SplitHostPort(scb.client.Host)
	if err != nil {
		return nil, err
	}

	return scb.client, nil
}

type smtpClient struct {
	client   *jordanEmail.Email
	Host     string
	Login    string
	Password string
	TLS      bool
	SSL      bool
}

func (sc *smtpClient) GetHostname() string {
	hostname, _, _ := net.SplitHostPort(sc.Host)
	return hostname
}

func (sc *smtpClient) Send(em *email) (err error) {
	if err = sc.checkParam(em); err != nil {
		return
	}

	sc.client.From = em.From
	if sc.client.From == "" {
		sc.client.From = sc.Login
	}

	sc.client.To = em.To
	sc.client.Cc = em.CarbonCopy
	sc.client.Bcc = em.BlindCarbonCopy

	sc.client.Subject = em.Subject

	if em.Text != nil {
		sc.client.Text = em.Text.Bytes()
	}

	if em.HTML != nil {
		sc.client.HTML = em.HTML.Bytes()
	}

	if em.Attachments != nil {
		err = sc.processAttach(em)
		if err != nil {
			return
		}
	}

	switch {
	case sc.TLS:
		err = sc.client.SendWithTLS(
			sc.Host,
			smtp.PlainAuth("", sc.Login, sc.Password, sc.GetHostname()),
			&tls.Config{ServerName: sc.Host},
		)

		if err != nil {
			return
		}
	case sc.SSL:
		err = sc.SendWithSSL()
		if err != nil {
			return
		}
	default:
		err = sc.client.Send(sc.Host, smtp.PlainAuth("", sc.Login, sc.Password, sc.GetHostname()))
		if err != nil {
			return
		}
	}

	return nil
}

func (sc *smtpClient) SendWithSSL() (err error) {
	to, sender, rawMessage, err := sc.getSendData()
	if err != nil {
		return
	}

	conn, err := tls.Dial("tcp", sc.Host, &tls.Config{ServerName: sc.GetHostname()})
	if err != nil {
		return
	}

	c, err := smtp.NewClient(conn, sc.Host)
	if err != nil {
		return
	}

	// Auth
	if err = c.Auth(smtp.PlainAuth("", sc.Login, sc.Password, sc.Host)); err != nil {
		return
	}

	// To && From
	if err = c.Mail(sender); err != nil {
		return
	}

	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return
	}

	_, err = w.Write(rawMessage)
	if err != nil {
		return
	}

	err = w.Close()
	if err != nil {
		return
	}

	err = c.Quit()
	if err != nil {
		return
	}

	return err
}

func (sc *smtpClient) getSendData() (to []string, sender string, rawMessage []byte, err error) {
	to, err = sc.getTo()
	if err != nil {
		return
	}

	// Check to make sure there is at least one recipient and one "From" address
	if sc.client.From == "" || len(to) == 0 {
		err = errors.New("must specify at least one From address and one To address")
		return
	}

	sender, err = sc.parseSender()
	if err != nil {
		return
	}

	rawMessage, err = sc.client.Bytes()
	if err != nil {
		return
	}

	return
}

func (sc *smtpClient) getTo() ([]string, error) {
	// Merge the To, Cc, and Bcc fields
	to := make([]string, 0, len(sc.client.To)+len(sc.client.Cc)+len(sc.client.Bcc))
	to = append(append(append(to, sc.client.To...), sc.client.Cc...), sc.client.Bcc...)

	for i := 0; i < len(to); i++ {
		addr, err := mail.ParseAddress(to[i])
		if err != nil {
			return nil, err
		}

		to[i] = addr.Address
	}

	return to, nil
}

// Select and parse an SMTP envelope sender address.  Choose Email.Sender if set, or fallback to Email.From.
func (sc *smtpClient) parseSender() (string, error) {
	if sc.client.Sender != "" {
		sender, err := mail.ParseAddress(sc.client.Sender)
		if err != nil {
			return "", err
		}

		return sender.Address, nil
	}

	from, err := mail.ParseAddress(sc.client.From)
	if err != nil {
		return "", err
	}

	return from.Address, nil
}

func (sc *smtpClient) checkParam(em *email) (err error) {
	if sc.Host == "" {
		return errors.New(emptyHost)
	}

	if em.To == nil || len(em.To) == 0 {
		return errors.New(needRecipient)
	}

	for _, e := range em.To {
		if e == "" {
			err = errors.New(emptyRecipient)
			return
		}
	}

	if em.Subject == "" && em.Text == nil && em.HTML == nil && (em.Attachments == nil || len(em.Attachments) == 0) {
		return errors.New(emptyData)
	}

	return
}

func (sc *smtpClient) processAttach(e *email) (err error) {
	for _, attach := range e.Attachments {
		if attach.filename == "" {
			err = errors.New(emptyAttachFilename)
			return
		}

		if attach.reader != nil || attach.data != nil {
			if attach.contentType == "" {
				err = errors.New(emptyAttachContentType)
				return
			}
		}

		switch {
		case attach.reader != nil:
			if _, err = sc.client.Attach(attach.reader, attach.filename, attach.contentType); err != nil {
				return
			}
		case attach.data != nil && len(attach.data) > 0:
			buf := bytes.NewBuffer(attach.data)
			if _, err = sc.client.Attach(buf, attach.filename, attach.contentType); err != nil {
				return
			}
		default:
			_, err = sc.client.AttachFile(attach.filename)
			if err != nil {
				return
			}
		}
	}

	return err
}
