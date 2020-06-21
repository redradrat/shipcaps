package v1beta1

import (
	"context"
	"fmt"
	"strings"

	helmv1 "github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
	"github.com/oliveagle/jsonpath"
	"github.com/redradrat/shipcaps/errors"
	"github.com/redradrat/shipcaps/parsing"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (app *App) ParentCap(client client.Client, ctx context.Context) (interface{ GetSpec() CapSpec }, error) {

	if app.Spec.CapRef.IsClusterCap() {
		var outcap Cap

		if err := client.Get(ctx, app.Spec.CapRef.NamespacedName(), &outcap); err != nil {
			return nil, err
		}

		return &outcap, nil

	} else {

		var outcap ClusterCap

		if err := client.Get(ctx, app.Spec.CapRef.NamespacedName(), &outcap); err != nil {
			return nil, err
		}

		return &outcap, nil

	}
}

// CreateOrUpdate creates/updates the actual App on the cluster,
// minus additional data input possibilities
func (app *App) CreateOrUpdate(c client.Client, ctx context.Context) (map[string]string, error) {
	return app.CreateOrUpdateWithData(c, ctx)
}

// CreateOrUpdate creates/updates the actual App on the cluster
func (app *App) CreateOrUpdateWithData(c client.Client, ctx context.Context, data ...map[string]string) (map[string]string, error) {

	outputMap := make(map[string]string)

	// First, we get the Parent Cap for this App. This is essential so we can actually map our App.
	// If the referenced Cap isn't found, there is no need to go further.
	parentCap, err := app.ParentCap(c, ctx)
	if err != nil {
		return outputMap, err
	}

	// Second, we render the CapValues. So if we did not get proper AppValues, we can already err here.
	capValues, err := app.RenderInputs(parentCap, c, ctx, data...)
	if err != nil {
		return outputMap, err
	}

	// Now we actually create/update our App
	err = app.CreateOrUpdateRecursively(parentCap, capValues, c, ctx)
	if err != nil {
		return outputMap, err
	}

	// Finally we render our outputs, and create our output secret
	outputMap, err = app.RenderOutputs(parentCap, c, ctx)
	if err != nil {
		return outputMap, err
	}

	return outputMap, nil
}

// RenderValues takes an App Object as input and uses its spec to render a complete set of CapValues
func (app *App) RenderInputs(parentCap interface{ GetSpec() CapSpec }, c client.Client, ctx context.Context, data ...map[string]string) (map[string]interface{}, error) {
	var outValues map[string]interface{}

	// Unmarshal the Values from our Cap and put them onto the output map
	cvs, err := parsing.ParseRawCapValues(parsing.RawCapValues(parentCap.GetSpec().Values))
	if err != nil {
		return nil, err
	}
	for _, cv := range cvs {
		outValues[string(cv.TargetIdentifier)] = cv.Value
	}

	// Now we put the given additional data onto our target map, and thus replace any
	// potentially already existing values. -> Additional Data takes precedence over CapValues
	for _, dataSet := range data {
		for k, v := range dataSet {
			outValues[k] = v
		}
	}

	// Finally we get the App Values, given as Inputs. These take precedence over everything.
	// Unmarshal given App's values.
	avs, err := parsing.ParseRawAppValues(parsing.RawAppValues(app.Spec.Values))
	if err != nil {
		return nil, err
	}

	// Go through the whole map and see if all Inputs are given, and have the right type.
	// Then add them to our output map, replacing potentially existing keys there.
	avMap := avs.Map()
	for _, in := range parentCap.GetSpec().Inputs {
		var err bool
		data, found := avMap[in.Key]
		if !found {
			if !in.Optional {
				return nil, fmt.Errorf("required key '%s' not found in App values", in.Key)
			}
			continue
		}
		switch in.Type {
		case StringInputType:
			if _, ok := data.(string); !ok {
				err = true
			}
		case StringListInputType:
			if _, ok := data.([]string); !ok {
				err = true
			}
		case IntInputType:
			if _, ok := data.(int); !ok {
				err = true
			}
		case FloatInputType:
			if _, ok := data.(float32); !ok {
				err = true
			}
		}
		if err {
			return nil, fmt.Errorf("required input '%s' is not of type '%s'", in.Key, in.Type)
		}
		// Value looks good, let's put it onto our output map.
		outValues[string(in.TargetIdentifier)] = data
	}

	return outValues, nil
}

func (app *App) RenderOutputs(parentCap interface{ GetSpec() CapSpec }, c client.Client, ctx context.Context) (map[string]string, error) {
	outputMap := make(map[string]string)

	for _, output := range parentCap.GetSpec().Outputs {
		unstruct := unstructured.Unstructured{}
		err := c.Get(ctx, client.ObjectKey{Name: output.ObjectRef.Name, Namespace: output.ObjectRef.Namespace}, &unstruct)
		if err != nil {
			return outputMap, err
		}

		outputVal, err := jsonpath.JsonPathLookup(unstruct.UnstructuredContent, output.FieldRef.FieldPath)
		if err != nil {
			return outputMap, err
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

	_, err := ctrl.CreateOrUpdate(ctx, c, &secret, mutateFunction)
	if err != nil {
		return outputMap, err
	}

	return outputMap, nil
}

func (app *App) CreateOrUpdateSimple(src CapSource, capValues map[string]interface{}, c client.Client, ctx context.Context, refs ...v1.OwnerReference) error {

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
		_, err := ctrl.CreateOrUpdate(ctx, c, &entry, couFunc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *App) CreateOrUpdateHelm(src CapSource, capValues map[string]interface{}, c client.Client, ctx context.Context) error {

	helmValueMap := makeHelmValues(capValues)

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
	_, err := ctrl.CreateOrUpdate(ctx, c, &helmRel, couFunc)
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

const (
	InvalidCapSourceType errors.ShipCapsErrorCode = "InvalidCapSourceType"
)

func (app *App) CreateOrUpdateRecursively(parentCap interface{ GetSpec() CapSpec }, capValues map[string]interface{}, c client.Client, ctx context.Context) error {

	sourceType := parentCap.GetSpec().Source.Type

	switch sourceType {
	case SimpleCapSourceType:
		if err := app.CreateOrUpdateSimple(parentCap.GetSpec().Source, capValues, c, ctx); err != nil {
			return err
		}
	case HelmChartCapSourceType:
		if err := app.CreateOrUpdateHelm(parentCap.GetSpec().Source, capValues, c, ctx); err != nil {
			return err
		}
	}

	return errors.NewShipCapsError(InvalidCapSourceType, fmt.Sprintf("unknown CapSource Type '%s'", sourceType))
}
