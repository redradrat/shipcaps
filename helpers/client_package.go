package helpers

import (
	"context"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClientPackage struct {
	Client  client.Client
	Context context.Context
	Logger  *logr.Logger
	Owner   metav1.Object
	Scheme  *runtime.Scheme
}

func (cp *ClientPackage) Get(key client.ObjectKey, obj runtime.Object) error {
	return cp.Client.Get(cp.Context, key, obj)
}

func (cp *ClientPackage) List(obj runtime.Object, opts ...client.ListOption) error {
	return cp.Client.List(cp.Context, obj, opts...)
}

func (cp *ClientPackage) Create(obj runtime.Object, opts ...client.CreateOption) error {
	return cp.Client.Create(cp.Context, obj, opts...)
}

func (cp *ClientPackage) Delete(obj runtime.Object, opts ...client.DeleteOption) error {
	return cp.Client.Delete(cp.Context, obj, opts...)
}

func (cp *ClientPackage) SetControllerReference(obj metav1.Object) {
	controllerruntime.SetControllerReference(cp.Owner, obj, cp.Scheme)
}
