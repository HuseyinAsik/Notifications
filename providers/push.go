package providers

type PushProvider struct{}

func (p *PushProvider) Send(id, to, content string) error {
	return nil
}
