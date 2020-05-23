package v1beta1

import (
	"github.com/redradrat/shipcaps/helpers"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cap *Cap) ResolveVersion(pkg helpers.ClientPackage) (*CapVersion, error) {
	capversions := CapVersionList{}
	pkg.List(&capversions, client.MatchingFields{
		"spec.capRef.name":      cap.Name,
		"spec.capRef.namespace": cap.Namespace,
		"spec.capRef.kind":      cap.Kind,
	})
	return &CapVersion{}, nil
}

func (cap *ClusterCap) ResolveVersion(pkg helpers.ClientPackage) (*CapVersion, error) {
	capversions := CapVersionList{}
	pkg.List(&capversions, client.MatchingFields{
		"spec.capRef.name": cap.Name,
		"spec.capRef.kind": cap.Kind,
	})
	capversion
	return &CapVersion{}, nil
}
