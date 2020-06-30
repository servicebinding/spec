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

package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ServiceBindingApplication defines an extension to v1.ObjectReference
type ServiceBindingApplication struct {
	v1.ObjectReference `json:",inline"`
	// Containers describes which containers in a Pod should be bound to
	Containers []intstr.IntOrString `json:"containers,omitempty"`
	// Selector is a query that selects the application or applications to bind the service to
	Selector metav1.LabelSelector `json:"selector,omitempty"`
}

// ServiceBindingEnvVar defines a mapping from the value of a Secret entry to an environment variable
type ServiceBindingEnvVar struct {
	// Name is the name of the environment variable
	Name string `json:"name"`
	// Key is the key in the Secret that will be exposed
	Key string `json:"key"`
}

// ServiceBindingMapping defines a mapping from the existing collection of Secret values to a new Secret entry.
type ServiceBindingMapping struct {
	// Name is the name of the mapped Secret entry
	Name string `json:"name"`
	// Value is the value of the new Secret entry.  Contents may be a Go template and refer to the other secret entries
	// by name.
	Value string `json:"value"`
}

// ServiceBindingSpec defines the desired state of ServiceBinding
type ServiceBindingSpec struct {
	// Name is the name of the service as projected into the application container.  Defaults to .metadata.name.
	Name string `json:"name,omitempty"`
	// Type is the type of the service as projected into the application container
	Type string `json:"type,omitempty"`
	// Provider is the provider of the service as projected into the application container
	Provider string `json:"provider,omitempty"`
	// Application is a reference to an object that fulfills the PodSpec duck type
	Application ServiceBindingApplication `json:"application"`
	// Service is a reference to an object that fulfills the ProvisionedService duck type
	Service v1.ObjectReference `json:"service"`
	// EnvVars is the collection of mappings from Secret entries to environment variables
	EnvVars []ServiceBindingEnvVar `json:"env,omitempty"`
	// Mappings is the collection of mappings from existing Secret entries to new Secret entries
	Mappings []ServiceBindingMapping `json:"mappings,omitempty"`
}

// ServiceBindingConditionType is a valid value for ServiceBindingCondition.Type
type ServiceBindingConditionType string

// These are valid conditions of ServiceBinding.
const (
	// ServiceBindingReady means the ServiceBinding has projected the ProvisionedService secret and the Pod is ready to
	// start
	ServiceBindingReady ServiceBindingConditionType = "Ready"
)

// ServiceBindingCondition contains details for the current condition of this ServiceBinding
type ServiceBindingCondition struct {
	// Type is the type of the condition
	Type ServiceBindingConditionType `json:"type"`
	// Status is the status of the condition
	// Can be True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// Last time the condition transitioned from one status to another
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition
	Message string `json:"message,omitempty"`
}

// ServiceBindingStatus defines the observed state of ServiceBinding
type ServiceBindingStatus struct {

	// +kubebuilder:validation:MinItems=1

	// Conditions are the conditions of this ServiceBinding
	Conditions []ServiceBindingCondition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ServiceBinding is the Schema for the servicebindings API
type ServiceBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceBindingSpec   `json:"spec,omitempty"`
	Status ServiceBindingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceBindingList contains a list of ServiceBinding
type ServiceBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ServiceBinding `json:"items"`
}
