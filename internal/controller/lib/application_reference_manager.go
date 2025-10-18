package lib

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/bensoer/rebasecd/api/v1alpha1"
)

type ApplicationReferenceManager struct {
	application *operatorv1alpha1.Application
	client      client.Client
}

func NewApplicationReferenceManager(application *operatorv1alpha1.Application, client client.Client) *ApplicationReferenceManager {
	return &ApplicationReferenceManager{
		client:      client,
		application: application,
	}
}

func (asm *ApplicationReferenceManager) UpdateApplicationReference(ctx context.Context, namespaceName client.ObjectKey) error {
	return asm.client.Get(ctx, namespaceName, asm.application)
}
