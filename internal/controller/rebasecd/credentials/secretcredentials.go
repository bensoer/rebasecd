package credentials

import (
	"context"
	"fmt"

	rebasecd "github.com/bensoer/rebasecd/internal/controller/rebasecd"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretCredentials struct {
	SecretName      string
	SecretNamespace string
	UsernameKey     string
	PasswordKey     string

	client client.Client
}

func NewSecretCredentials(secretName, secretNamespace, usernameKey, passwordKey string, client client.Client) rebasecd.CredentialsHandler {
	return &SecretCredentials{
		SecretName:      secretName,
		SecretNamespace: secretNamespace,
		UsernameKey:     usernameKey,
		PasswordKey:     passwordKey,
		client:          client,
	}
}

func (sc *SecretCredentials) GetUsername() string {
	// Implementation to retrieve username from secret
	username, err := sc.getSecretKeyValue(context.Background(), sc.SecretName, sc.UsernameKey)
	if err != nil {
		return ""
	}
	return username
}

func (sc *SecretCredentials) GetPassword() string {
	// Implementation to retrieve password from secret
	password, err := sc.getSecretKeyValue(context.Background(), sc.SecretName, sc.PasswordKey)
	if err != nil {
		return ""
	}
	return password
}

func (sc *SecretCredentials) HasCredentials() bool {
	return sc.SecretName != "" && sc.UsernameKey != "" && sc.PasswordKey != ""
}

func (sc *SecretCredentials) getSecretKeyValue(ctx context.Context, secretName, key string) (string, error) {
	// Retrieve the Secret
	secret := &corev1.Secret{}
	err := sc.client.Get(ctx, client.ObjectKey{Namespace: sc.SecretNamespace, Name: secretName}, secret)
	if err != nil {
		return "", fmt.Errorf("failed to get secret %q in namespace %q: %w", secretName, sc.SecretNamespace, err)
	}

	// Extract the key from the Secret
	valueBytes, ok := secret.Data[key]
	if !ok {
		return "", fmt.Errorf("key %q not found in secret %q", key, secretName)
	}

	return string(valueBytes), nil
}
