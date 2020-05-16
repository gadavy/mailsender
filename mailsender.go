package mailsender

const (
	emptyAttachFilename    = "empty filename"
	emptyAttachContentType = "empty content-type"
	emptyHost              = "empty host"
	needRecipient          = "need recipient"
	emptyRecipient         = "empty recipient"
	emptyData              = "empty data"
)

type Builder interface {
	Email() EmailBuilder
	SMTPClient() SMTPClientBuilder
}

func New() Builder {
	return &builder{}
}

type builder struct{}

func (b *builder) SMTPClient() SMTPClientBuilder {
	return NewSMTPClient()
}

func (b *builder) Email() EmailBuilder {
	return NewEmail()
}
