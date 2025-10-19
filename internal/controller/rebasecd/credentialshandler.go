package rebasecd

type CredentialsHandler interface {
	GetUsername() string
	GetPassword() string
	HasCredentials() bool
}
