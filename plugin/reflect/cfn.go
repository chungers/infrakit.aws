package reflect

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type AWSClients struct {
	Cfn cloudformationiface.CloudFormationAPI
	Ec2 ec2iface.EC2API
	Asg autoscalingiface.AutoScalingAPI
}

type cfnPlugin struct {
	clients       AWSClients
	namespaceTags map[string]string
}

// NewCFNPlugin creates a new plugin that can introspect a Cloudformation stack
func NewCFNPlugin(clients AWSClients, namespaceTags map[string]string) Plugin {
	return &cfnPlugin{clients: clients, namespaceTags: namespaceTags}
}

func (c *cfnPlugin) Render(model EnvironmentModel, templateURL string) (string, error) {
	t, err := NewTemplate(templateURL)
	if err != nil {
		return "", err
	}

	t.AddFunc("describe", func(p string, o interface{}) interface{} {
		obj := get(o, tokenize(p))
		r, is := obj.(*cloudformation.StackResource)
		if !is || r == nil {
			return nil
		}
		d, err := describe(c.clients, r)
		if err == nil {
			return d
		}
		return err
	})

	return t.Render(model)
}

func (c *cfnPlugin) Inspect(name string) (EnvironmentModel, error) {
	model := EnvironmentModel{}

	input := cloudformation.DescribeStacksInput{
		StackName: &name,
	}

	output, err := c.clients.Cfn.DescribeStacks(&input)
	if err != nil {
		return model, err
	}

	if len(output.Stacks) == 0 {
		return model, fmt.Errorf("invalid stack %v", name)
	}

	output2, err := c.clients.Cfn.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
		StackName: &name,
	})
	if err != nil {
		return model, err
	}

	// index resources by type/name
	resources := map[string]map[string]interface{}{}
	for _, r := range output2.StackResources {
		if r.ResourceType == nil {
			continue
		}
		if r.LogicalResourceId == nil {
			continue
		}

		if resources[*r.ResourceType] == nil {
			resources[*r.ResourceType] = map[string]interface{}{}
		}
		resources[*r.ResourceType][*r.LogicalResourceId] = r
	}
	model.Resources = resources

	// index parameters by name
	parameters := map[string]interface{}{}
	for _, p := range output.Stacks[0].Parameters {
		if p.ParameterKey == nil {
			continue
		}
		parameters[*p.ParameterKey] = p
	}
	model.Parameters = parameters

	return model, nil
}
