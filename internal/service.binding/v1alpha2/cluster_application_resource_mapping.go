/*
 * Copyright 2021 the original author or authors.
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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// ClusterApplicationResourceMappingVersion defines the mapping for a specific version of an application resource.
type ClusterApplicationResourceMappingVersion struct {
	// Version is the version of the application resource that this mapping is for.
	Version string `json:"version"`
	// Containers is the collection of JSONPaths that container configuration may be written to.
	Containers []string `json:"containers,omitempty"`
	// Envs is the collection of JSONPaths that env configuration may be written to.
	Envs []string `json:"envs,omitempty"`
	// VolumeMounts is the collection of JSONPaths that volume mount configuration may be written to.
	VolumeMounts []string `json:"volumeMounts,omitempty"`
	// Volumes is the JSONPath that volume configuration must be written to.
	Volumes string `json:"volumes"`
}

// ClusterApplicationResourceMappingSpec defines the desired state of ClusterApplicationResourceMapping
type ClusterApplicationResourceMappingSpec struct {
	// Versions is the collection of versions for a given resource, with mappings.
	Versions []ClusterApplicationResourceMappingVersion `json:"versions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ClusterApplicationResourceMapping is the Schema for the clusterapplicationresourcemappings API
type ClusterApplicationResourceMapping struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterApplicationResourceMappingSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterApplicationResourceMappingList contains a list of ClusterApplicationResourceMapping
type ClusterApplicationResourceMappingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ClusterApplicationResourceMapping `json:"items"`
}
