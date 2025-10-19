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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type HelmChartCredentialsSpec struct {

	// Explicitly pass the username for authenticating with the Helm Chart repository
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:example:string:"john.smith@gmail.com
	Username string `json:"username,omitempty"`

	// Explicitly pass the password for authenticating with the Helm Chart repository
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:example:string:
	Password string `json:"password,omitempty"`

	// Specify a secret that contains teh username and password for authenticating with the Helm Chart Repository
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	SecretName string `json:"secretName,omitempty"`

	// The Key within the secret that contains the username for authenticating with the Helm Chart Repository
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	UsernameKey string `json:"usernameKey,omitempty"`

	// The Key within the secret that contains the password for authenticating with the Helm Chart Repository
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	PasswordKey string `json:"passwordKey,omitempty"`
}

// ApplicationSpec defines the desired state of Application.
type ApplicationSpec struct {

	// How often (in minutes) to poll the Helm Chart repository for changes
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=3
	// +kubebuilder:validation:Minimum=1
	PollIntervalMinutes int32 `json:"pollIntervalMinutes,omitempty"`

	// The URL where the Helm Chart is located
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Required
	Url string `json:"url,omitempty"`

	// Name of the Helm Chart. Only used if URL is pointing to a Helm Chart Repository
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Optional
	// +kubebuilder:example:string:"my-service-chart"
	ChartName string `json:"chartName,omitempty"`

	// For git Helm charts, the path within the repository where the chart is located
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Optional
	// +kubebuilder:example:string:"/chart"
	// +kubebuilder:default:"/chart"
	Path string `json:"path,omitempty"`

	// The Helm Chart Version
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Required
	Version string `json:"version,omitempty"`

	// The Values Files within the Helm Chart to use with this Application
	// +kubebuilder:validation:Required
	ValuesFiles []string `json:"valuesFiles,omitempty"`

	// Release name for the Helm Chart Deployment
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Required
	// +kubebuilder:example:string:"my-service-release"
	ReleaseName string `json:"releaseName,omitempty"`

	// The namespace where the Helm Chart will be deployed to. Defaults to the "default" namespace
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:"default"
	// +kubebuilder:validation:Type=string
	// +kubebuilder:example:string:"serviceA"
	TargetNamespace string `json:"targetNamespace,omitempty"`

	// The Credentials for authenticating with the Helm Chart Repository
	// +optional
	HelmChartCredentials HelmChartCredentialsSpec `json:"helmChartCredentials,omitempty"`
}

// ApplicationStatus defines the observed state of Application.
type ApplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Application is the Schema for the applications API.
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationList contains a list of Application.
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
