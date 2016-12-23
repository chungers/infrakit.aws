{{ with env "STACK_NAME" | cfn }}

{{/* ref a data structure and assign it as an alias */}}
{{ ref "/Resources/AWS::EC2::Subnet/PubSubnetAz1" . | alias "binding" }}
{
   "VPCId" : "{{ ref "/Resources/AWS::EC2::VPC/Vpc/PhysicalResourceId" . }}",
   "SubnetId" : "{{ ref "/Resources/AWS::EC2::Subnet/PubSubnetAz1/PhysicalResourceId" . }}",
   "Managers" : {{ ref "/Parameters/ManagerSize/ParameterValue" . }},
   "ManagerAsgBlockDevice" : "{{ describe "/Resources/AWS::AutoScaling::LaunchConfiguration/ManagerLaunchConfigBeta13" . | ref "/BlockDeviceMappings[0]/DeviceName" }}",
   "PubSubnetAz1Cidr" : "{{ describe "/Resources/AWS::EC2::Subnet/PubSubnetAz1" . | ref "/CidrBlock" }}",
   "VpcCidrBlock" : "{{ describe "/Resources/AWS::EC2::VPC/Vpc" . | ref "/CidrBlock"}}",
   "ManagerAsg" : {{ describe "/Resources/AWS::AutoScaling::AutoScalingGroup/ManagerAsg" . | jsonEncode }},
   "ManagerLaunch" : {{ describe "/Resources/AWS::AutoScaling::LaunchConfiguration/ManagerLaunchConfigBeta13" . | jsonEncode }},
   "Include" : {{ include "include.tpl" . }},

   {{/* references a value that's been aliased with 'as' */}}
   "binding" : {{ var "binding" "some binding object" | jsonEncode }}
}

{{ end }}
