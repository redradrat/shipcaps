package webhooks

import (
	"context"
	"fmt"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/redradrat/shipcaps/api/v1beta1"
)

// +kubebuilder:webhook:path=/validate-v1beta1-app,mutating=false,failurePolicy=fail,groups="shipcaps.redradrat.xyz",resources=apps,verbs=create;update,versions=v1,name=vapp.shipcaps.redradrat.xyz

const AppValidatorPath = "/validate-v1beta1-app"

type AppValidator struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (v *AppValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	app := &v1beta1.App{}
	err := v.decoder.Decode(req, app)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	cap := &v1beta1.Cap{}
	err = v.decoder.Decode(req, cap)
	if err != nil {
		return admission.ValidationResponse(
			false,
			fmt.Sprintf("referenced Cap '%s/%s' does not exist", cap.Namespace, cap.Name))
	}

	// Let's check if all required inputs from the Cap are in our App
	// So let's execute the Value rendering and see whether we get any errors there
	if _, err := cap.RenderValues(app); err != nil {
		return admission.ValidationResponse(
			false,
			err.Error())
	}

	// Now we know that we did set everything properly
	return admission.ValidationResponse(true, "all required key for referenced cap were provided")
}

func (a *AppValidator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}
