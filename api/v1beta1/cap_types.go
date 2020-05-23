/*


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

package v1beta1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CapSpec defines the desired state of Cap
type CapSpec struct {

	// Versions allows for the selection of CapVersions that this Cap exposes
	//
	// +kubebuilder:validation:Required
	Versions metav1.LabelSelector `json:"versions,omitempty"`

	// Dependencies specify Apps that this App depends on
	//
	// +kubebuilder:validation:Optional
	Dependencies []v1.ObjectReference `json:"dependencies,omitempty"`
}

// CapStatus defines the observed state of Cap
type CapStatus struct {
	// +kubebuilder:validation:optional
	//
	// ObservedGeneration holds the generation (metadata.generation in CR) observed by the controller
	ObservedGeneration int64 `json:"observedGeneration"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=caps,shortName=cap

// Cap is the Schema for the caps API
type Cap struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CapSpec   `json:"spec,omitempty"`
	Status CapStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CapList contains a list of Cap
type CapList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cap `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cap{}, &CapList{})
}
