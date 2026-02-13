package providers

type EmailProvider struct{}

func (e *EmailProvider) Send(id, to, content string) error {
	return nil
}
