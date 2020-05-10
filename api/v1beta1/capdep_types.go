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
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CapDepSpec defines the desired state of CapDep
type CapDepSpec struct {
	// Values allows to specify provided values. This can reduce user choice when using a Helm Chart for example.
	//
	// +kubebuilder:validation:Optional
	Values json.RawMessage `json:"values,omitempty"`

	// Source is an object reference to the required CapSource
	//
	// +kubebuilder:validation:Required
	Source CapSource `json:"source"`
}

// CapDepStatus defines the observed state of CapDep
type CapDepStatus struct {
	// +kubebuilder:validation:optional
	//
	// ObservedGeneration holds the generation (metadata.generation in CR) observed by the controller
	ObservedGeneration int64 `json:"observedGeneration"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=capdeps

// CapDep is the Schema for the capdeps API
type CapDep struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CapDepSpec   `json:"spec,omitempty"`
	Status CapDepStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CapDepList contains a list of CapDep
type CapDepList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CapDep `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CapDep{}, &CapDepList{})
}
