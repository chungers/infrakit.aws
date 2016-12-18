package reflect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/docker/infrakit/pkg/manager"
)

type Plugin interface {
	Inspect(name string) (manager.GlobalSpec, error)
}

const (
	// VolumeTag is the AWS tag name used to associate unique identifiers (instance.VolumeID) with volumes.
	VolumeTag = "docker-infrakit-volume"
)

type cfnPlugin struct {
	client        cloudformationiface.CloudFormationAPI
	namespaceTags map[string]string
}

// NewCFNPlugin creates a new plugin that can introspect a Cloudformation stack
func NewCFNPlugin(client cloudformationiface.CloudFormationAPI, namespaceTags map[string]string) Plugin {
	return &cfnPlugin{client: client, namespaceTags: namespaceTags}
}

func (c *cfnPlugin) Inspect(name string) (manager.GlobalSpec, error) {
	spec := manager.GlobalSpec{}

	log.Infoln("Inspecting", name)

	input := cloudformation.DescribeStacksInput{
		StackName: &name,
	}

	output, err := c.client.DescribeStacks(&input)
	if err != nil {
		return spec, err
	}

	if len(output.Stacks) == 0 {
		return spec, fmt.Errorf("invalid stack %v", name)
	}

	output2, err := c.client.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
		StackName: &name,
	})
	if err != nil {
		return spec, err
	}

	combined := map[string]interface{}{}

	// index resources by type/name
	resources := map[string]interface{}{}
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
		resources[*r.ResourceType].(map[string]interface{})[*r.LogicalResourceId] = r
	}
	combined["Resources"] = resources

	// index parameters by name
	parameters := map[string]interface{}{}
	for _, p := range output.Stacks[0].Parameters {
		if p.ParameterKey == nil {
			continue
		}
		parameters[*p.ParameterKey] = p
	}
	combined["Parameters"] = parameters

	// index outputs by name
	outputs := map[string]interface{}{}
	for _, o := range output.Stacks[0].Outputs {
		if o.OutputKey == nil {
			continue
		}
		outputs[*o.OutputKey] = o
	}
	combined["Outputs"] = outputs
	l := []interface{}{}
	for _, s := range output2.StackResources {
		l = append(l, s)
	}
	combined["ResourcesList"] = l

	// dump out json
	buff, err := json.MarshalIndent(combined, "", "  ")
	if err != nil {
		return spec, err
	}
	log.Infoln(string(buff))

	// apply template

	test := `
{
   "VPCId" : "{{ ref . "/Resources/AWS::EC2::VPC/Vpc/PhysicalResourceId" }}",
   "LastResourceId" : "{{ ref . "/ResourcesList[-1]/PhysicalResourceId" }}",
   "SubnetId" : "{{ ref . "/Resources/AWS::EC2::Subnet/PubSubnetAz1/PhysicalResourceId" }}",
   "Managers" : {{ ref . "/Parameters/ManagerSize/ParameterValue" }}
}
`
	t, err := template.New("template").Funcs(map[string]interface{}{
		"ref": func(o interface{}, p string) interface{} {
			return get(o, tokenize(p))
		},
	}).Parse(test)
	if err != nil {
		return spec, err
	}

	var buffer bytes.Buffer
	err = t.Execute(&buffer, combined)
	if err != nil {
		return spec, err
	}

	log.Infoln(buffer.String())
	return spec, nil
}
