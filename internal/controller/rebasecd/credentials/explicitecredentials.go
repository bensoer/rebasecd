package credentials

import rebasecd "github.com/bensoer/rebasecd/internal/controller/rebasecd"

type ExpliciteCredentials struct {
	Username string
	Password string
}

func NewExpliciteCredentials(username, password string) rebasecd.CredentialsHandler {
	return &ExpliciteCredentials{
		Username: username,
		Password: password,
	}
}

func (ec *ExpliciteCredentials) GetUsername() string {
	return ec.Username
}

func (ec *ExpliciteCredentials) GetPassword() string {
	return ec.Password
}

func (ec *ExpliciteCredentials) HasCredentials() bool {
	return ec.Username != "" && ec.Password != ""
}
