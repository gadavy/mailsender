package mailsender

import (
	"crypto/tls"
	"errors"
	"net/smtp"
)

type loginAuth struct {
	username, password, domain string
}

func LoginAuth(username, password, domain string) smtp.Auth {
	return &loginAuth{username, password, domain}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("unknown from server")
		}
	}
	return nil, nil
}

func (sc *smtpClient) SendWithTLS() (err error) {
	to, sender, rawMessage, err := sc.getSendData()
	if err != nil {
		return
	}

	c, err := smtp.Dial(sc.Host)
	if err != nil {
		return
	}

	err = c.StartTLS(&tls.Config{ServerName: sc.GetHostname()})
	if err != nil {
		return
	}

	// Auth
	if err = c.Auth(LoginAuth(sc.Login, sc.Password, sc.GetHostname())); err != nil {
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
