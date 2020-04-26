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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CapType specifies the type of a Cap. Used for identifying a backend.
type CapType string

const (
	// SimpleCapType is a simple value-replacement Cap
	SimpleCapType CapType = "simple"

	// HelmChartCapType abstracts a Helm Chart as a Cap
	HelmChartCapType CapType = "helmchart"
)

// InputType specifies the type of an Input. Used for parsing.
type InputType string

const (
	// StringInputType identifies an Input should be parsed as string
	StringInputType InputType = "string"
	// PasswordInputType identifies an Input should be parsed as string but handled with care https://www.youtube.com/watch?v=1o4s1KVJaVA
	PasswordInputType InputType = "password"
	// IntInputType identifies an Input should be parsed as int
	IntInputType InputType = "int"
	// FloatInputType identifies an Input should be parsed as float
	FloatInputType InputType = "float"
	// StringListInputType identifies an Input should be parsed as a list of string
	StringListInputType InputType = "stringlist"
)

// CapInput defines an Input required for our Cap
type CapInput struct {
	Key string `json:"key"`

	// Type identifies the type of the this input (string, int, ...). Used for parsing.
	//
	// +kubebuilder:validation:Required
	Type InputType `json:"type"`

	// Optional identifies whether this Input is required or not
	//
	// +kubebuilder:validation:Optional
	Optional bool `json:"optional"`
}

// CapInputs is a list of CapInputs
type CapInputs []CapInput

// MaterialAuth references authentication credentials for a Helm Chart Repo
type MaterialAuth struct {
	// Username is the username to authenticate with for the Repository
	//
	// +kubebuilder:validation:Required
	Username v1.EnvVarSource `json:"username"`

	// Password is the password to authenticate with for the Repository
	//
	// +kubebuilder:validation:Required
	Password v1.EnvVarSource `json:"password"`
}

// RepoSpec specifies a specific git repository and revision
type RepoSpec struct {
	// RepoURI is the URI to the specific git repository (GitOps y'all!)
	//
	// +kubebuilder:validation:Required
	URI string `json:"uri"`

	// Ref specifies a GitRef to use for the repo
	//
	// +kubebuilder:validation:Optional
	Ref string `json:"ref,omitempty"`

	// Auth potentially needed authentication credentials for the referenced material
	//
	// +kubebuilder:validation:Optional
	Auth MaterialAuth `json:"auth,omitempty"`
}

// HelmMaterial is the helm-specific material
type HelmMaterial struct {
}

// HelmMaterial is the simple-specific material
type SimpleMaterial struct {
}

// CapMaterial is a the material of the respective supported types
type CapMaterial struct {

	// Repo is specification of the git repository (GitOps y'all!)
	//
	// +kubebuilder:validation:Required
	Repo RepoSpec `json:"repo"`

	// Path specifies a subpath in the repo specified via repo
	//
	// +kubebuilder:validation:Optional
	Path string `json:"path,omitempty"`

	// Manifests holds a list of manifests to use as material
	//
	// +kubebuilder:validation:XEmbeddedResource
	// +kubebuilder:validation:Optional
	Manifests []unstructured.Unstructured `json:"manifests,omitempty"`

	Helm HelmMaterial `json:",inline"`

	Simple SimpleMaterial `json:",inline"`

	Plain PlainMaterial `json:",inline"`
}

// CapSpec defines the desired state of Cap
type CapSpec struct {

	// Type specifies the type of to our Cap (e.g. what is our backend? Helm, Manifests, ...)
	//
	// +kubebuilder:validation:Required
	Type CapType `json:"type"`

	// Inputs specify all Inputs that can be given to our Cap
	//
	// +kubebuilder:validation:Optional
	Inputs CapInputs `json:"inputs,omitempty"`

	// Matter specifies the matter of the specified type (helm chart, manifests, etc.)
	//
	// +kubebuilder:validation:Required
	Material CapMaterial `json:"material"`

	// Dependencies specify Apps that this App depends on
	//
	// +kubebuilder:validation:Optional
	Dependencies v1.ObjectReference `json:"dependencies,omitempty"`
}

// CapStatus defines the observed state of Cap
type CapStatus struct {
	// +kubebuilder:validation:optional
	//
	// ObservedGeneration holds the generation (metadata.generation in CR) observed by the controller
	ObservedGeneration int64 `json:"observedGeneration"`
}

// +kubebuilder:object:root=true

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
