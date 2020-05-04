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
	"strconv"
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

	log.V(1).Info(fmt.Sprintf("Reconciling app '%s/%s'", app.Name, app.Namespace))

	cap := shipcapsv1beta1.Cap{}
	err = r.Client.Get(ctx, client.ObjectKey{Name: app.Spec.CapRef}, &cap)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := cap.Spec.Material.Check(); err != nil {
		return ctrl.Result{}, err
	}

	capValues, err := MergedCapValues(cap, app, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	switch cap.Spec.Type {
	case shipcapsv1beta1.SimpleCapType:
		if err := r.ReconcileSimpleCapTypeApp(cap, app, capValues, ctx, log); err != nil {
			return ctrl.Result{}, err
		}
	case shipcapsv1beta1.HelmChartCapType:
		if err := r.ReconcileHelmChartCapTypeApp(cap, app, capValues, ctx, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{
		RequeueAfter: r.RequeueDuration,
	}, nil
}

func (r *AppReconciler) ReconcileHelmChartCapTypeApp(cap shipcapsv1beta1.Cap, app shipcapsv1beta1.App, capValues []CapValue, ctx context.Context, log logr.Logger) error {
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
			GitURL: cap.Spec.Material.Repo.URI,
			Ref:    cap.Spec.Material.Repo.Ref,
			Path:   cap.Spec.Material.Repo.Path,
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

func (r *AppReconciler) ReconcileSimpleCapTypeApp(cap shipcapsv1beta1.Cap, app shipcapsv1beta1.App, capValues []CapValue, ctx context.Context, log logr.Logger) error {
	var processedOut []unstructured.Unstructured

	mat := cap.Spec.Material
	switch mat.Type {
	case shipcapsv1beta1.ManifestsMaterialType:
		newbytes := SubSimplePlaceholders(mat.Manifests, capValues)
		if PlaceholdersLeft(newbytes) {
			return fmt.Errorf("not all required values have been given")
		}
		newparsed := []unstructured.Unstructured{}
		if err := json.Unmarshal(newbytes, &newparsed); err != nil {
			return err
		}
		processedOut = append(processedOut, newparsed...)
	default:
		return fmt.Errorf("material type '%s' not yet supported for cap type '%s'", mat.Type, cap.Spec.Type)
	}

	for _, entry := range processedOut {

		couFunc := func() error { return nil }
		if entry.GetNamespace() != "" {
			if err := controllerutil.SetControllerReference(&app, &entry, r.Scheme); err != nil {
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

func SubSimplePlaceholders(in []byte, vals []CapValue) []byte {
	out := string(in)
	for _, val := range vals {
		switch val.Value.(type) {
		case string:
			out = strings.ReplaceAll(string(out), fmt.Sprintf("!Rep{%s}", val.TargetIdentifier), val.Value.(string))
		case int:
			out = strings.ReplaceAll(string(out), fmt.Sprintf("!Rep{%s}", val.TargetIdentifier), strconv.Itoa(val.Value.(int)))
		case float32:
			out = strings.ReplaceAll(string(out), fmt.Sprintf("!Rep{%s}", val.TargetIdentifier), fmt.Sprintf("%.2f", val.Value.(float32)))
		case []string:
			joinedVal := strings.Join(val.Value.([]string), ", ")
			out = strings.ReplaceAll(string(out), fmt.Sprintf("!Rep{%s}", val.TargetIdentifier), joinedVal)
		}
	}
	return []byte(out)
}

func PlaceholdersLeft(in []byte) bool {
	if strings.Contains(string(in), "!Rep{") {
		return true
	}
	return false
}

func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&shipcapsv1beta1.App{}).
		Owns(&helmv1.HelmRelease{}).
		Complete(r)
}
