package credentials

import (
	operatorv1alpha1 "github.com/bensoer/rebasecd/api/v1alpha1"
	rebasecd "github.com/bensoer/rebasecd/internal/controller/rebasecd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateCredentialsFromSpec(spec *operatorv1alpha1.HelmChartCredentialsSpec, namespace string, client client.Client) (rebasecd.CredentialsHandler, error) {

	if spec.Username != "" && spec.Password != "" {
		return NewExpliciteCredentials(spec.Username, spec.Password), nil
	} else if spec.SecretName != "" && spec.UsernameKey != "" && spec.PasswordKey != "" {
		return NewSecretCredentials(spec.SecretName, namespace, spec.UsernameKey, spec.PasswordKey, client), nil
	} else {
		return NewEmptyCredentials(), nil
	}

}
