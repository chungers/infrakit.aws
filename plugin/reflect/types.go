package reflect

// EnvironmentModel is a common model that captures information about a cloud environment.
// It can be a version of CloudFormation data or Azure resource templates
type EnvironmentModel struct {

	// Resources is an index of resource by type name, then name.
	Resources []interface{}

	// Parameters is an index of user-defined parameters.
	Parameters []interface{}
}

// Plugin defines the possible interfactions with the reflection service
type Plugin interface {
	// Render the template at the given URL to produce a global spec for InfraKit.
	Render(templateURL string, context interface{}) (string, error)
}
