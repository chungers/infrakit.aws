package reflect

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

var (
	ErrNotFound = fmt.Errorf("not-found")
)

var asgFuncs = map[string]func(clients AWSClients, r *cloudformation.StackResource) (interface{}, error){
	"LaunchConfiguration": func(clients AWSClients, r *cloudformation.StackResource) (interface{}, error) {

		fmt.Println(">>>>", *r.LogicalResourceId)
		resp, err := clients.Asg.DescribeLaunchConfigurations(&autoscaling.DescribeLaunchConfigurationsInput{
			LaunchConfigurationNames: []*string{r.PhysicalResourceId},
		})
		if err != nil {
			return nil, err
		}
		if len(resp.LaunchConfigurations) == 0 {
			return nil, ErrNotFound
		}
		return resp.LaunchConfigurations[0], nil
	},
}

func describe(clients AWSClients, r *cloudformation.StackResource) (interface{}, error) {
	resourceType := *r.ResourceType
	i := strings.LastIndex(resourceType, "::")
	t := resourceType[i+2:]
	switch resourceType[0:i] {
	case "AWS::EC2":
		fmt.Println("Looking up ec2", t)
	case "AWS::AutoScaling":
		fmt.Println("Looking up asg", t)
		return asgFuncs[t](clients, r)

	default:
		return nil, nil
	}
	return r, nil
}
