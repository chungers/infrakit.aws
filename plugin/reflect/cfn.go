package reflect

import (
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
	clients AWSClients
}

// NewCFNPlugin creates a new plugin that can introspect a Cloudformation stack
func NewCFNPlugin(clients AWSClients) Plugin {
	return &cfnPlugin{clients: clients}
}

func (c *cfnPlugin) Render(templateURL string, context interface{}) (string, error) {
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
	t.AddFunc("cfn", func(p string) (interface{}, error) {
		return cfn(c.clients, p)
	})
	return t.Render(context)
}
