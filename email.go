package mailsender

import (
	"bytes"
	"io"
)

type EmailBuilder interface {
	Subject(subject string) EmailBuilder
	From(sender string) EmailBuilder
	To(recipients ...string) EmailBuilder
	CarbonCopy(copyes ...string) EmailBuilder
	BlindCarbonCopy(blindCopyes ...string) EmailBuilder
	Text(text string) EmailBuilder
	HTML(html string) EmailBuilder
	Attach(filename, contectType string, data []byte) EmailBuilder
	AttachFromReader(filename, contectType string, r io.Reader) EmailBuilder
	AttachFromFile(filename string) EmailBuilder
	Build() *email
}

type Email interface {
	WithTo(recipients ...string) Email
	WithCarbonCopy(copyes ...string) Email
	WithBlindCarbonCopy(blindCopyes ...string) Email
}

func NewEmail() EmailBuilder {
	return &emailBuilder{
		email: &email{},
	}
}

type emailBuilder struct {
	email *email
}

func (eb *emailBuilder) Subject(subj string) EmailBuilder {
	eb.email.Subject = subj
	return eb
}

func (eb *emailBuilder) From(from string) EmailBuilder {
	eb.email.From = from
	return eb
}

func (eb *emailBuilder) To(to ...string) EmailBuilder {
	eb.email.To = to
	return eb
}

func (eb *emailBuilder) CarbonCopy(cc ...string) EmailBuilder {
	eb.email.CarbonCopy = cc
	return eb
}

func (eb *emailBuilder) BlindCarbonCopy(bcc ...string) EmailBuilder {
	eb.email.BlindCarbonCopy = bcc
	return eb
}

func (eb *emailBuilder) Text(text string) EmailBuilder {
	eb.email.Text = bytes.NewBufferString(text)
	return eb
}

func (eb *emailBuilder) HTML(html string) EmailBuilder {
	eb.email.HTML = bytes.NewBufferString(html)
	return eb
}

// Attach - params: Filename, MIME Content-Type, data
func (eb *emailBuilder) Attach(filename, contentType string, data []byte) EmailBuilder {
	if eb.email.Attachments == nil {
		eb.email.Attachments = make([]attach, 0, 1)
	}

	eb.email.Attachments = append(eb.email.Attachments, attach{
		filename:    filename,
		contentType: contentType,
		data:        data,
	})

	return eb
}

// AttachFromReader - params: Filename, MIME Content-Type, data
func (eb *emailBuilder) AttachFromReader(filename, contentType string, r io.Reader) EmailBuilder {
	if eb.email.Attachments == nil {
		eb.email.Attachments = make([]attach, 0, 1)
	}

	eb.email.Attachments = append(eb.email.Attachments, attach{
		filename:    filename,
		contentType: contentType,
		reader:      r,
	})

	return eb
}

// AttachFromFile - params: Filename, MIME Content-Type, data
func (eb *emailBuilder) AttachFromFile(filename string) EmailBuilder {
	if eb.email.Attachments == nil {
		eb.email.Attachments = make([]attach, 0, 1)
	}

	eb.email.Attachments = append(eb.email.Attachments, attach{
		filename: filename,
	})

	return eb
}

func (eb *emailBuilder) Build() *email {
	return eb.email
}

type email struct {
	Subject         string
	From            string
	To              []string
	CarbonCopy      []string
	BlindCarbonCopy []string
	Text            *bytes.Buffer
	HTML            *bytes.Buffer
	Attachments     []attach
}

type attach struct {
	filename    string
	contentType string
	data        []byte
	reader      io.Reader
}

func (e *email) WithTo(recipients ...string) Email {
	e.To = append(e.To, recipients...)
	return e
}

func (e *email) WithCarbonCopy(copyes ...string) Email {
	e.CarbonCopy = append(e.CarbonCopy, copyes...)
	return e
}

func (e *email) WithBlindCarbonCopy(copyes ...string) Email {
	e.BlindCarbonCopy = append(e.BlindCarbonCopy, copyes...)
	return e
}
