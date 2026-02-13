package providers

type SMSProvider struct{}

func (s *SMSProvider) Send(id, to, content string) error {

	return nil
}
