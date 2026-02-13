package providers

type Provider interface {
	Send(id, to string, content string) error
}
