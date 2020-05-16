package mailsender

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmail(t *testing.T) {

	em := &email{
		Subject:         "subj",
		Text:            bytes.NewBufferString("text"),
		HTML:            bytes.NewBufferString("html"),
		From:            "from@localhost",
		To:              []string{"to@localhost"},
		CarbonCopy:      []string{"cc@localhost"},
		BlindCarbonCopy: []string{"bcc@localhost"},
		Attachments: []attach{
			attach{
				filename:    "test.txt",
				contentType: "text/plain",
				data:        []byte("text"),
			},
			attach{
				filename: "test2.txt",
			},
			attach{
				filename:    "test3.txt",
				contentType: "text/plain",
				reader:      bytes.NewBuffer([]byte("text2")),
			},
		},
	}

	em2 := New().Email().
		Subject(em.Subject).
		Text(em.Text.String()).
		HTML(em.HTML.String()).
		From(em.From).
		To(em.To...).
		CarbonCopy(em.CarbonCopy...).
		BlindCarbonCopy(em.BlindCarbonCopy...).
		Attach("test.txt", "text/plain", []byte("text")).
		AttachFromFile("test2.txt").
		AttachFromReader("test3.txt", "text/plain", bytes.NewBuffer([]byte("text2"))).
		Build()

	assert.Equal(t, em, em2)

	em.WithTo("to2@localhost")
	assert.Equal(t, em.To, []string{"to@localhost", "to2@localhost"})

	em.WithCarbonCopy("cc2@localhost")
	assert.Equal(t, em.CarbonCopy, []string{"cc@localhost", "cc2@localhost"})

	em.WithBlindCarbonCopy("bcc2@localhost")
	assert.Equal(t, em.BlindCarbonCopy, []string{"bcc@localhost", "bcc2@localhost"})
}
