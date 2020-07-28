/*
 * Copyright 2020 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ServiceBindingProjectionSecretReference defines a mirror of corev1.LocalObjectReference
type ServiceBindingProjectionSecretReference struct {
	// Name of the referent secret.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string `json:"name"`
}

// ServiceBindingProjectionApplicationReference defines a subset of corev1.ObjectReference with extensions
type ServiceBindingProjectionApplicationReference struct {
	// API version of the referent.
	APIVersion string `json:"apiVersion"`
	// Kind of the referent.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	Kind string `json:"kind"`
	// Name of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string `json:"name,omitempty"`
	// Selector is a query that selects the application or applications to bind the service to
	Selector metav1.LabelSelector `json:"selector,omitempty"`
	// Containers describes which containers in a Pod should be bound to
	Containers []intstr.IntOrString `json:"containers,omitempty"`
}

// ServiceBindingProjectionEnvVar defines a mapping from the value of a Secret entry to an environment variable
type ServiceBindingProjectionEnvVar struct {
	// Name is the name of the environment variable
	Name string `json:"name"`
	// Key is the key in the Secret that will be exposed
	Key string `json:"key"`
}

// ServiceBindingProjectionSpec defines the desired state of ServiceBindingProjection
type ServiceBindingProjectionSpec struct {
	// Name is the name of the service as projected into the application container.
	Name string `json:"name"`
	// Binding is the projected secret for this ServiceBindingProjection.
	Binding ServiceBindingProjectionSecretReference `json:"binding"`
	// Application is a reference to an object that fulfills the PodSpec duck type
	Application ServiceBindingProjectionApplicationReference `json:"application"`
	// EnvVars is the collection of mappings from Secret entries to environment variables
	EnvVars []ServiceBindingProjectionEnvVar `json:"env,omitempty"`
}

// ServiceBindingProjectionConditionType is a valid value for ServiceBindingProjectionCondition.Type
type ServiceBindingProjectionConditionType string

// These are valid conditions of ServiceBindingProjection.
const (
	// ServiceBindingProjectionReady means the ServiceBindingProjection has projected the ProvisionedService secret and
	//the Pod is ready to start
	ServiceBindingProjectionReady ServiceBindingProjectionConditionType = "Ready"
)

// ServiceBindingProjectionCondition contains details for the current condition of this ServiceBindingProjection
type ServiceBindingProjectionCondition struct {
	// Type is the type of the condition
	Type ServiceBindingProjectionConditionType `json:"type"`
	// Status is the status of the condition
	// Can be True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// Last time the condition transitioned from one status to another
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition
	Message string `json:"message,omitempty"`
}

// ServiceBindingProjectionStatus defines the observed state of ServiceBindingProjection
type ServiceBindingProjectionStatus struct {
	// ObservedGeneration is the 'Generation' of the ServiceBindingProjection that
	// was last processed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions are the conditions of this ServiceBindingProjection
	Conditions []ServiceBindingProjectionCondition `json:"conditions,omitempty"`

}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].reason`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ServiceBindingProjection is the Schema for the servicebindingprojections API
type ServiceBindingProjection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceBindingProjectionSpec   `json:"spec,omitempty"`
	Status ServiceBindingProjectionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceBindingProjectionList contains a list of ServiceBindingProjection
type ServiceBindingProjectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ServiceBindingProjection `json:"items"`
}
