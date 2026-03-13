/*
Copyright 2026.

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

// Package v1 contains API Schema definitions for the example v1 API group
// +kubebuilder:object:generate=true
// +groupName=example.com
// +kubebuilder:storageversion
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// JsonServerSpec defines the desired state of JsonServer.
type JsonServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Replicas are a number of instances of JsonServer
	// +kubebuilder:validation:int32
	// +kubebuilder:validation:Minimum=0
	Replicas int32 `json:"replicas,omitempty"`

	//JsonConfig represents config passed to be processed by JsonServer
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:MinLength=1
	JsonConfig string `json:"jsonConfig"`
}

// JsonServerStatus defines the observed state of JsonServer.
type JsonServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Message regarsing state
	Message string `json:"message,omitempty"`

	// Current State of JsonServer
	State State `json:"state,omitempty"`
}

type State string

const (
	StateSynced State = "Synced"
	StateError  State = "Error"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector
// +kubebuilder:webhook:path=/validate-example-com-v1-JsonServer,mutating=false,failurePolicy=fail,sideEffects=None,groups=example.com,resources=JsonServers,verbs=create;update,versions=v1,name=vJsonServer.kb.io,admissionReviewVersions=v1
// JsonServer is the Schema for the JsonServers API.
type JsonServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JsonServerSpec   `json:"spec,omitempty"`
	Status JsonServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// JsonServerList contains a list of JsonServer.
type JsonServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JsonServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JsonServer{}, &JsonServerList{})
}
