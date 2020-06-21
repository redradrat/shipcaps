package requirements

type AppRequirement struct {
	name               string
	namespace          string
	globalDependencies []*AppRequirement
	uniqueDependencies []*AppRequirement
}

func New(name, namespace string, globalDependencies, uniqueDependencies []*AppRequirement) *AppRequirement {
	return &AppRequirement{
		name:               name,
		namespace:          namespace,
		globalDependencies: globalDependencies,
		uniqueDependencies: uniqueDependencies,
	}
}

func (app *AppRequirement) ComputeGlobalDependencies() error {
	return nil
}

func (app *AppRequirement) ComputeUniqueDependencies() error {
	return nil
}
