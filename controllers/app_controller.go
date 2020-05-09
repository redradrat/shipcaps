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

	cap := shipcapsv1beta1.Cap{}
	err = r.Client.Get(ctx, client.ObjectKey{Name: app.Spec.CapRef}, &cap)
	if err != nil {
		return ctrl.Result{}, err
	}

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

func (r *AppReconciler) ReconcileHelmChartCapTypeApp(src shipcapsv1beta1.CapSource, app shipcapsv1beta1.App, capValues parsing.CapValues, ctx context.Context, log logr.Logger) error {
	helmValueMap := make(map[string]interface{})
	for _, val := range capValues {
		helmValueMap[string(val.TargetIdentifier)] = val.Value
	}

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
