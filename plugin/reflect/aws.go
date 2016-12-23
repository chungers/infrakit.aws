package reflect

import (
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var (
	ErrNotFound     = fmt.Errorf("not-found")
	ErrNotSupported = fmt.Errorf("not-supported")
)

type call struct {
	method   interface{}
	input    interface{}
	selector string
}

func doDescribe(r *cloudformation.StackResource, c call) (interface{}, error) {
	method := reflect.ValueOf(c.method)
	if method.IsNil() {
		return nil, ErrNotSupported
	}
	resp := method.Call([]reflect.Value{reflect.ValueOf(c.input)})

	var err error
	if !resp[1].IsNil() {
		err = resp[1].Interface().(error)
	}

	if err != nil {
		return nil, err
	}

	return get(resp[0].Interface(), tokenize(c.selector)), err
}

var describeFuncs = map[string]func(AWSClients, *cloudformation.StackResource) (interface{}, error){
	"AWS::AutoScaling::LaunchConfiguration": func(clients AWSClients, r *cloudformation.StackResource) (interface{}, error) {
		return doDescribe(r, call{
			method: clients.Asg.DescribeLaunchConfigurations,
			input: &autoscaling.DescribeLaunchConfigurationsInput{
				LaunchConfigurationNames: []*string{r.PhysicalResourceId},
			},
			selector: "/LaunchConfigurations[0]",
		})
	},
	"AWS::AutoScaling::AutoScalingGroup": func(clients AWSClients, r *cloudformation.StackResource) (interface{}, error) {
		return doDescribe(r, call{
			method: clients.Asg.DescribeAutoScalingGroups,
			input: &autoscaling.DescribeAutoScalingGroupsInput{
				AutoScalingGroupNames: []*string{r.PhysicalResourceId},
			},
			selector: "/AutoScalingGroups[0]",
		})
	},
	"AWS::EC2::Subnet": func(clients AWSClients, r *cloudformation.StackResource) (interface{}, error) {
		return doDescribe(r, call{
			method: clients.Ec2.DescribeSubnets,
			input: &ec2.DescribeSubnetsInput{
				SubnetIds: []*string{r.PhysicalResourceId},
			},
			selector: "/Subnets[0]",
		})
	},
	"AWS::EC2::VPC": func(clients AWSClients, r *cloudformation.StackResource) (interface{}, error) {
		return doDescribe(r, call{
			method: clients.Ec2.DescribeVpcs,
			input: &ec2.DescribeVpcsInput{
				VpcIds: []*string{r.PhysicalResourceId},
			},
			selector: "/Vpcs[0]",
		})
	},
}

func describe(clients AWSClients, r *cloudformation.StackResource) (interface{}, error) {
	resourceType := *r.ResourceType
	if f, has := describeFuncs[resourceType]; has {
		return f(clients, r)
	}
	return nil, ErrNotSupported
}

func cfn(clients AWSClients, name string) (EnvironmentModel, error) {
	model := EnvironmentModel{}

	input := cloudformation.DescribeStacksInput{
		StackName: &name,
	}

	output, err := clients.Cfn.DescribeStacks(&input)
	if err != nil {
		return model, err
	}

	if len(output.Stacks) == 0 {
		return model, fmt.Errorf("invalid stack %v", name)
	}

	output2, err := clients.Cfn.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
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
