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
	"fmt"
	"strings"
	"time"

	helmv1 "github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
	"github.com/go-logr/logr"
	"github.com/oliveagle/jsonpath"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	shipcapsv1beta1 "github.com/redradrat/shipcaps/api/v1beta1"
	"github.com/redradrat/shipcaps/parsing"
)

// AppReconciler reconciles a App object
type AppReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	RequeueDuration time.Duration
}

// +kubebuilder:rbac:groups=shipcaps.redradrat.xyz,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=shipcaps.redradrat.xyz,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=shipcaps.redradrat.xyz,resources=caps,verbs=get;list;watch
// +kubebuilder:rbac:groups=shipcaps.redradrat.xyz,resources=caps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=shipcaps.redradrat.xyz,resources=capdeps,verbs=get;list;watch
// +kubebuilder:rbac:groups=shipcaps.redradrat.xyz,resources=capdeps/status,verbs=get;update;patch

func (r *AppReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("app", req.NamespacedName)

	var app shipcapsv1beta1.App
	err := r.Get(ctx, req.NamespacedName, &app)
	if err != nil {
		log.V(1).Info("unable to fetch App")
		return ctrl.Result{
			RequeueAfter: r.RequeueDuration,
		}, client.IgnoreNotFound(err)
	}

	// Get the referenced Cap/ClusterCap
	_, err = app.ParentCap(r.Client, ctx)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("referenced Caps are unavailable: %s", err)
		} else {
			return ctrl.Result{}, err
		}
	}

	// Reconcile the defined dependencies
	// if err := r.reconcileDeps(app, parentCap, ctx); err != nil {
	// 	return ctrl.Result{}, err
	// }

	// Reconcile the actual App now
	_, err = app.CreateOrUpdate(r.Client, ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Done
	log.V(1).Info("Successfully Reconciled")
	return ctrl.Result{
		RequeueAfter: r.RequeueDuration,
	}, nil
}

// Creates the secret holding the outputs for the App
func (r *AppReconciler) createAppSecret(app shipcapsv1beta1.App, parentCap shipcapsv1beta1.Cap, ctx context.Context) error {

	outputMap := make(map[string]string)

	for _, output := range parentCap.Spec.Outputs {
		unstruct := unstructured.Unstructured{}
		err := r.Get(ctx, client.ObjectKey{Name: output.ObjectRef.Name, Namespace: output.ObjectRef.Namespace}, &unstruct)
		if err != nil {
			return err
		}

		outputVal, err := jsonpath.JsonPathLookup(unstruct.UnstructuredContent, output.FieldRef.FieldPath)
		if err != nil {
			return err
		}

		outputMap[output.TargetIdentifier] = outputVal.(string)
	}

	secret := corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
	}

	mutateFunction := func() error {
		secret.StringData = outputMap
		return nil
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &app, mutateFunction)
	if err != nil {
		return err
	}

	return nil

}

func makeHelmValues(in map[string]interface{}) map[string]interface{} {
	// create output map
	var out = make(map[string]interface{})

	// iterate through all input keys
	for k, v := range in {
		// separate key segments
		keysegments := strings.Split(k, ".")

		// new map var from out map. different reference, same underlying object
		inter := out

		// iterate through all segments -1 to create our map hierarchy; last one will be assigned directly
		for _, seg := range keysegments[:len(keysegments)-1] {
			// get the value if segment already exists
			new, ok := inter[seg]
			if !ok {
				// segment didn't exist, let's create a new map object and put it as value
				new = make(map[string]interface{})
				inter[seg] = new
			}
			// We will now overwrite our inter reference to be the new map object, as we want to
			// iterate deeper into the hierarchy.
			// Out will still reference the highest point of the underlying object.
			inter = new.(map[string]interface{})
		}

		// we can assign the value finally, as inter now references the
		// deepest map object in our hierarchy.
		inter[keysegments[len(keysegments)-1]] = v
	}
	return out

}

func (r *AppReconciler) ReconcileHelmChartCapTypeApp(src shipcapsv1beta1.CapSource, app shipcapsv1beta1.App, capValues parsing.CapValues, ctx context.Context, log logr.Logger) error {
	helmValueMap := makeHelmValues(capValues.Map())

	helmRel := helmv1.HelmRelease{
		ObjectMeta: v1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
	}
	couFunc := func() error {
		helmRel.Spec.Values = helmValueMap
		cs := helmv1.GitChartSource{
			GitURL: src.Repo.URI,
			Ref:    src.Repo.Ref,
			Path:   src.Repo.Path,
		}
		helmRel.Spec.GitChartSource = &cs
		return nil
	}
	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &helmRel, couFunc)
	if err != nil {
		return err
	}

	return nil
}

//func (r *AppReconciler) reconcileDeps(app shipcapsv1beta1.App, cap shipcapsv1beta1.Cap, ctx context.Context) error {
//	var resolvedDeps []shipcapsv1beta1.Cap
//
//	for _, dep := range cap.Spec.Dependencies {
//		var cap shipcapsv1beta1.Cap
//		var key client.ObjectKey
//
//		if dep.Namespace == "" {
//			key = client.ObjectKey{Name: dep.Name}
//		} else {
//			key = client.ObjectKey{Name: dep.Name, Namespace: dep.Namespace}
//		}
//
//		if err := r.Get(ctx, key, &cap); err != nil {
//			return err
//		}
//		resolvedDeps = append(resolvedDeps, cap)
//	}
//
//	// Reconcile the Dependencies for this App
//	for _, dep := range resolvedDeps {
//		depValues, err := dep.RenderValues()
//		if err != nil {
//			return err
//		}
//
//		switch dep.Spec.Source.Type {
//		case shipcapsv1beta1.SimpleCapSourceType:
//			if err := r.ReconcileSimpleCapTypeApp(dep.Spec.Source, app, depValues, ctx, log); err != nil {
//				return err
//			}
//		case shipcapsv1beta1.HelmChartCapSourceType:
//			if err := r.ReconcileHelmChartCapTypeApp(dep.Spec.Source, app, depValues, ctx, log); err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}

func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&shipcapsv1beta1.App{}).
		Owns(&helmv1.HelmRelease{}).
		Complete(r)
}
