package mailsender

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/emersion/go-smtp"
	jordanEmail "github.com/jordan-wright/email"

	"github.com/stretchr/testify/assert"
)

func TestSMTPClient(t *testing.T) {

	go runSMTPServer()
	time.Sleep(time.Millisecond * 500)

	sc := &smtpClient{
		Login:    "login@localhost",
		Password: "password",
		Host:     "localhost:8025",
		TLS:      false,
		client:   jordanEmail.NewEmail(),
	}

	sc2, err := New().SMTPClient().
		Login(sc.Login).
		Password(sc.Password).
		Host(sc.Host).
		TLS(sc.TLS).
		Build()

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, sc, sc2)

	err = sc.Send(New().Email().
		To("test@localhost").
		Subject("subj").
		Text("text").
		Attach("test.txt", "text/plain", []byte("data")).
		Build())
	if err != nil {
		t.Fatal(err)
	}

}

// The Backend implements SMTP server methods.
type Backend struct{}

// Login handles a login command with username and password.
func (bkd *Backend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	if username != "login@localhost" || password != "password" {
		return nil, errors.New("Invalid username or password")
	}
	return &Session{}, nil
}

// AnonymousLogin requires clients to authenticate using SMTP AUTH before sending emails
func (bkd *Backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	return nil, smtp.ErrAuthRequired
}

// A Session is returned after successful login.
type Session struct{}

func (s *Session) Mail(from string) error {
	//log.Println("Mail from:", from)
	return nil
}

func (s *Session) Rcpt(to string) error {
	//	log.Println("Rcpt to:", to)
	return nil
}

func (s *Session) Data(r io.Reader) error {
	if _, err := ioutil.ReadAll(r); err != nil {
		return err
	} else {
		//	log.Println("Data:", string(b))
	}
	return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error {
	return nil
}

func runSMTPServer() {
	be := &Backend{}

	s := smtp.NewServer(be)

	s.Addr = ":8025"
	s.Domain = "localhost"
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	s.AllowInsecureAuth = true

	//log.Println("Starting server at", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
