package reflect

// EnvironmentModel is a common model that captures information about a cloud environment.
// It can be a version of CloudFormation data or Azure resource templates
type EnvironmentModel struct {

	// Resources is an index of resource by type name, then name.
	Resources map[string]map[string]interface{}

	// Parameters is an index of user-defined parameters.
	Parameters map[string]interface{}
}

// Plugin defines the possible interfactions with the reflection service
type Plugin interface {
	// Inspect introspects the environment / stack by name.  For cloudformation, this is the stack name.
	Inspect(name string) (EnvironmentModel, error)

	// Render the template at the given URL to produce a global spec for InfraKit.
	Render(templateURL string) (string, error)
}
