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

type CapReference struct {
	// The name of the referred Cap
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// The namespace the referred Cap resides in (leave empty for cluster-scoped referenced)
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace"`
}

// AppSpec defines the desired state of App
type AppSpec struct {

	// CapRef refers to the Cap that should be applied
	//
	// +kubebuilder:validation:Required
	CapRef *CapReference `json:"capRef"`

	// Values is a list of inputs needed to create this app
	//
	// +kubebuilder:validation:Optional
	Values json.RawMessage `json:"values,omitempty"`
}

// AppStatus defines the observed state of App
type AppStatus struct {
	// +kubebuilder:validation:optional
	//
	// ObservedGeneration holds the generation (metadata.generation in CR) observed by the controller
	ObservedGeneration int64 `json:"observedGeneration"`
}

// +kubebuilder:object:root=true

// App is the Schema for the apps API
type App struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppSpec   `json:"spec,omitempty"`
	Status AppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AppList contains a list of App
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []App `json:"items"`
}

func init() {
	SchemeBuilder.Register(&App{}, &AppList{})
}
