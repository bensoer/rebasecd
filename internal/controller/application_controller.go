/*
Copyright 2025 bensoer.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	operatorv1alpha1 "github.com/bensoer/rebasecd/api/v1alpha1"
	"github.com/bensoer/rebasecd/internal/controller/lib"
)

const (
	// typeAvailableMemcached represents the status of the Deployment reconciliation
	typeAvailableApplication = "Available"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=operator.rebasecd.projectterris.com,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.rebasecd.projectterris.com,resources=applications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.rebasecd.projectterris.com,resources=applications/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Application object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	application := &operatorv1alpha1.Application{}
	arm := lib.NewApplicationReferenceManager(application, r.Client)

	// First check if the application definition exists. If it does not, stop reconciliation
	err := arm.UpdateApplicationReference(ctx, req.NamespacedName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Application Not Found. Ignore since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed To Get Application")
		return ctrl.Result{}, err
	}

	asm := lib.NewApplicationStatusManager(application, r.Client)

	// This means we found the application, so it exists
	// Lets check its status

	// If no status is available, set that status to Unknown
	if len(asm.GetCurrentConditions()) == 0 {

		err := asm.SetApplicationStatus(ctx, metav1.ConditionUnknown, "Reconciling", "Starting Reconciliation")
		if err != nil {
			log.Error(err, "Failed to Update Application Status")
			return ctrl.Result{}, err
		}

		// Fetch new version of the application data since changes were applied
		// This avoid conflict errors and us working with an older version
		err = arm.UpdateApplicationReference(ctx, req.NamespacedName)
		if err != nil {
			log.Error(err, "Failed to re-fetch Application")
			return ctrl.Result{}, err
		}
	}

	// This means it has a status, so lets start building the things

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Application{}).
		Named("application").
		Complete(r)
}
