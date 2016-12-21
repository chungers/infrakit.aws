{
   "VPCId" : "{{ ref "/Resources/AWS::EC2::VPC/Vpc/PhysicalResourceId" . }}",
   "SubnetId" : "{{ ref "/Resources/AWS::EC2::Subnet/PubSubnetAz1/PhysicalResourceId" . }}",
   "Managers" : {{ ref "/Parameters/ManagerSize/ParameterValue" . }},
   "ManagerAsgBlockDevice" : "{{ describe "/Resources/AWS::AutoScaling::LaunchConfiguration/ManagerLaunchConfigBeta13" . | ref "/BlockDeviceMappings[0]/DeviceName" }}",
   "PubSubnetAz1Cidr" : "{{ describe "/Resources/AWS::EC2::Subnet/PubSubnetAz1" . | ref "/CidrBlock" }}",
   "VpcCidrBlock" : "{{ describe "/Resources/AWS::EC2::VPC/Vpc" . | ref "/CidrBlock"}}",
   "ManagerAsg" : {{ describe "/Resources/AWS::AutoScaling::AutoScalingGroup/ManagerAsg" . | json }},
   "ManagerLaunch" : {{ describe "/Resources/AWS::AutoScaling::LaunchConfiguration/ManagerLaunchConfigBeta13" . | json }}
}
