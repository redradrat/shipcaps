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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redradrat/shipcaps/parsing"
)

// ValueType specifies the type of an Input. Used for parsing.
type ValueType string

const (
	// StringInputType identifies an Input should be parsed as string
	StringInputType ValueType = "string"
	// PasswordInputType identifies an Input should be parsed as string but handled with care https://www.youtube.com/watch?v=1o4s1KVJaVA
	PasswordInputType ValueType = "password"
	// IntInputType identifies an Input should be parsed as int
	IntInputType ValueType = "int"
	// FloatInputType identifies an Input should be parsed as float
	FloatInputType ValueType = "float"
	// StringListInputType identifies an Input should be parsed as a list of string
	StringListInputType ValueType = "stringlist"
)

// CapInput defines an Input required for our Cap
type CapInput struct {
	Key string `json:"key"`

	// Type identifies the type of the this input (string, int, ...). Used for parsing.
	//
	// +kubebuilder:validation:Required
	Type ValueType `json:"type"`

	// Optional identifies whether this Input is required or not
	//
	// +kubebuilder:validation:Optional
	Optional bool `json:"optional"`

	// TransformationIdentifier identifies the replacement placeholder.
	//
	// +kubebuilder:validation:Optional
	TargetIdentifier parsing.TargetIdentifier `json:"targetId,omitempty"`
}

// CapInputs is a list of CapInputs
type CapInputs []CapInput

// RepoAuth references authentication credentials for a Helm Chart Repo
type RepoAuth struct {
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

	// Path specifies a subpath in the repo
	//
	// +kubebuilder:validation:Optional
	Path string `json:"path,omitempty"`

	// Auth potentially needed authentication credentials for the referenced material
	//
	// +kubebuilder:validation:Optional
	Auth RepoAuth `json:"auth,omitempty"`
}

// CapSourceType specifies the type of a Cap. Used for identifying a backend.
type CapSourceType string

const (
	// SimpleCapSourceType is a simple value-replacement Cap
	SimpleCapSourceType CapSourceType = "simple"

	// HelmChartCapSourceType abstracts a Helm Chart as a Cap
	HelmChartCapSourceType CapSourceType = "helmchart"
)

type CapSource struct {
	// Type specifies the type of to our Cap (e.g. what is our backend? Helm, Manifests, ...)
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=helmchart;simple
	Type CapSourceType `json:"type"`

	// Repo is specification of the git repository (GitOps y'all!)
	//
	// +kubebuilder:validation:Optional
	Repo RepoSpec `json:"repo"`

	// InLine holds a list of manifests to use as material
	//
	// +kubebuilder:validation:Optional
	InLine json.RawMessage `json:"inline,omitempty"`
}

// CapSpec defines the desired state of Cap
type CapSpec struct {
	// Inputs specify all Inputs that can be given to our Cap
	//
	// +kubebuilder:validation:Optional
	Inputs CapInputs `json:"inputs,omitempty"`

	// Values allows to specify provided values. This can reduce user choice when using a Helm Chart for example.
	//
	// +kubebuilder:validation:Optional
	Values json.RawMessage `json:"values,omitempty"`

	// Source is an object reference to the required CapSource
	//
	// +kubebuilder:validation:Required
	Source CapSource `json:"source"`

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
// +kubebuilder:resource:path=caps,scope=Cluster,shortName=cap

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
