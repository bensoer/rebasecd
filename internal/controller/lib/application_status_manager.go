package lib

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/bensoer/rebasecd/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ApplicationStatusManager struct {
	application *operatorv1alpha1.Application
	client      client.Client
}

func NewApplicationStatusManager(application *operatorv1alpha1.Application, client client.Client) *ApplicationStatusManager {
	return &ApplicationStatusManager{
		client:      client,
		application: application,
	}
}

func (asm *ApplicationStatusManager) GetCurrentConditions() []v1.Condition {
	return asm.application.Status.Conditions
}

func (asm *ApplicationStatusManager) SetApplicationStatus(ctx context.Context, status metav1.ConditionStatus, reason string, message string) error {

	meta.SetStatusCondition(
		&asm.application.Status.Conditions,
		metav1.Condition{
			Type:    "Available",
			Status:  status,
			Reason:  reason,
			Message: message,
		},
	)

	err := asm.client.Status().Update(ctx, asm.application)
	if err != nil {
		return fmt.Errorf("failed to update application status %w", err)
	}

	return nil

}
