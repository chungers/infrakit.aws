package reflect

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/docker/infrakit/pkg/discovery"
	"github.com/docker/infrakit/pkg/template"
)

type AWSClients struct {
	Cfn cloudformationiface.CloudFormationAPI
	Ec2 ec2iface.EC2API
	Asg autoscalingiface.AutoScalingAPI
}

type cfnPlugin struct {
	templateOpts template.Options
	clients      AWSClients
}

// NewCFNPlugin creates a new plugin that can introspect a Cloudformation stack
func NewCFNPlugin(clients AWSClients) Plugin {
	c := &cfnPlugin{clients: clients, templateOpts: template.Options{SocketDir: discovery.Dir()}}
	return c
}

func (c *cfnPlugin) Render(templateURL string, context interface{}) (string, error) {
	t, err := template.NewTemplate(templateURL, c.templateOpts)
	if err != nil {
		return "", err
	}

	t.AddFunc("describe",
		func(p string, obj interface{}) (interface{}, error) {
			if obj == nil {
				return nil, nil
			}
			o, err := template.QueryObject(p, obj)
			if err != nil {
				return nil, err
			}

			switch o := o.(type) {

			case *cloudformation.StackResource:
				return describe(c.clients, o)

			case map[string]interface{}:
				rr := &cloudformation.StackResource{}
				err := template.FromMap(o, rr)
				if err != nil {
					return nil, err
				}
				return describe(c.clients, rr)

			}
			return nil, fmt.Errorf("unknown object:", o)
		})
	t.AddFunc("cfn", func(p string) (interface{}, error) {
		return cfn(c.clients, p)
	})

	return t.Render(context)
}
