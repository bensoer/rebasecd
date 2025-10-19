package credentials

import rebasecd "github.com/bensoer/rebasecd/internal/controller/rebasecd"

type EmptyCredentials struct {
}

func NewEmptyCredentials() rebasecd.CredentialsHandler {
	return &EmptyCredentials{}
}

func (ec *EmptyCredentials) GetUsername() string {
	return ""
}

func (ec *EmptyCredentials) GetPassword() string {
	return ""
}

func (ec *EmptyCredentials) HasCredentials() bool {
	return false
}
