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

package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	shipcapsv1beta1 "github.com/redradrat/shipcaps/api/v1beta1"
)

// CapReconciler reconciles a Cap object
type CapReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=shipcaps.redradrat.xyz,resources=caps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=shipcaps.redradrat.xyz,resources=caps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=shipcaps.redradrat.xyz,resources=capdeps,verbs=get;list;watch
// +kubebuilder:rbac:groups=shipcaps.redradrat.xyz,resources=capsdeps/status,verbs=get;update;patch

func (r *CapReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("cap", req.NamespacedName)

	var cap shipcapsv1beta1.Cap
	if err := r.Get(ctx, req.NamespacedName, &cap); err != nil {
		log.V(1).Info("unable to fetch Cap")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// return if only status/metadata updated
	if cap.Status.ObservedGeneration == cap.ObjectMeta.Generation {
		return ctrl.Result{}, nil
	} else {
		cap.Status.ObservedGeneration = cap.ObjectMeta.Generation
		if err := r.Status().Update(ctx, &cap); err != nil {
			return ctrl.Result{}, err
		}
	}

	src := cap.Spec.Source

	if src.IsInLine() {
		mans := src.InLine
		unstruct := []unstructured.Unstructured{}
		if err := json.Unmarshal(mans, &unstruct); err != nil {
			return ctrl.Result{}, err
		}
		for _, man := range unstruct {
			fmt.Printf("Resource: %s | Name: %s", man.GroupVersionKind().String(), man.GetName())
			//if err := r.Client.Create(ctx, &man, client.DryRunAll); err != nil {
			//	return ctrl.Result{}, err
			//}
		}
	}

	return ctrl.Result{}, nil
}

func (r *CapReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&shipcapsv1beta1.Cap{}).
		Complete(r)
}
