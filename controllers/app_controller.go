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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

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

	if app.Spec.ClusterCapRef != nil && app.Spec.CapRef != nil {
		return ctrl.Result{}, fmt.Errorf("both ClusterCapRef and CapRef set")
	}
	if app.Spec.ClusterCapRef == nil && app.Spec.CapRef == nil {
		return ctrl.Result{}, fmt.Errorf("neither ClusterCapRef nor CapRef set")
	}

	// Get the referenced ClusterCap
	var cap shipcapsv1beta1.Cap
	if app.Spec.ClusterCapRef != nil {
		clusterCap := shipcapsv1beta1.ClusterCap{}
		key := client.ObjectKey{
			Name: app.Spec.ClusterCapRef.Name,
		}
		err = r.Client.Get(ctx, key, &clusterCap)
		if err != nil {
			return ctrl.Result{}, err
		}
		cap = shipcapsv1beta1.Cap(clusterCap)
	}

	// Get the referenced Cap
	if app.Spec.CapRef != nil {
		key := client.ObjectKey{
			Namespace: app.Spec.CapRef.Namespace,
			Name:      app.Spec.CapRef.Name,
		}
		err = r.Client.Get(ctx, key, &cap)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// Get the required CapDeps
	var capdeps []shipcapsv1beta1.CapDep
	for _, dep := range cap.Spec.Dependencies {
		capdep := shipcapsv1beta1.CapDep{}
		err = r.Client.Get(ctx, client.ObjectKey{Name: dep.Name, Namespace: dep.Namespace}, &capdep)
		if err != nil {
			return ctrl.Result{}, err
		}
		capdeps = append(capdeps, capdep)
	}

	// Reconcile the Dependencies for this App
	for _, dep := range capdeps {
		depValues, err := dep.RenderValues()
		if err != nil {
			return ctrl.Result{}, err
		}

		switch dep.Spec.Source.Type {
		case shipcapsv1beta1.SimpleCapSourceType:
			if err := r.ReconcileSimpleCapTypeApp(dep.Spec.Source, &app, depValues, ctx, log); err != nil {
				return ctrl.Result{}, err
			}
		case shipcapsv1beta1.HelmChartCapSourceType:
			if err := r.ReconcileHelmChartCapTypeApp(dep.Spec.Source, app, depValues, ctx, log); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// Reconcile the App itself
	capValues, err := cap.RenderValues(&app)
	if err != nil {
		return ctrl.Result{}, err
	}

	switch cap.Spec.Source.Type {
	case shipcapsv1beta1.SimpleCapSourceType:
		if err := r.ReconcileSimpleCapTypeApp(cap.Spec.Source, &app, capValues, ctx, log); err != nil {
			return ctrl.Result{}, err
		}
	case shipcapsv1beta1.HelmChartCapSourceType:
		if err := r.ReconcileHelmChartCapTypeApp(cap.Spec.Source, app, capValues, ctx, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	log.V(1).Info("Successfully Reconciled")
	return ctrl.Result{
		RequeueAfter: r.RequeueDuration,
	}, nil
}

func makeHelmValues(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for key, val := range in {
		if strings.Contains(key, ".") {
			subs := strings.SplitN(key, ".", 2)
			out[subs[0]] = makeHelmValues(map[string]interface{}{subs[1]: val})
		} else {
			out[key] = val
		}
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

func (r *AppReconciler) ReconcileSimpleCapTypeApp(src shipcapsv1beta1.CapSource, app *shipcapsv1beta1.App, capValues parsing.CapValues, ctx context.Context, log logr.Logger) error {

	var err error
	if err = src.Check(); err != nil {
		return err
	}

	var processedOut unstructured.UnstructuredList
	if src.IsInLine() {
		processedOut, err = src.GetUnstructuredObjects(capValues)
		if err != nil {
			return err
		}
	}

	for _, entry := range processedOut.Items {
		couFunc := func() error { return nil }
		if entry.GetNamespace() != "" {
			if err := controllerutil.SetControllerReference(app, &entry, r.Scheme); err != nil {
				return err
			}
		}
		res, err := ctrl.CreateOrUpdate(ctx, r.Client, &entry, couFunc)
		log.V(1).Info(fmt.Sprintf("resource [kind: %s, name: %s, namespace: %s] %s", entry.GetKind(), entry.GetName(), entry.GetNamespace(), res))
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&shipcapsv1beta1.App{}).
		Owns(&helmv1.HelmRelease{}).
		Complete(r)
}
