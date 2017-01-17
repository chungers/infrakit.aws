{{ with env "STACK_NAME" | cfn }}

{
   "Dot" : {{ . | to_json }},
   "VPC" : {{ q "Resources[?ResourceType=='AWS::EC2::VPC'] | [0]" . | to_json}},
   "ManangerSize" : {{ q "Parameters[?ParameterKey=='ManagerSize'] | [0].ParameterValue" . }},
   "Vpc" : {{ describe "Resources[?ResourceType=='AWS::EC2::VPC'] | [0]" . | to_json}},
   "VpcCidrBlock" : {{ describe "Resources[?ResourceType=='AWS::EC2::VPC'] | [0]" . | q "CidrBlock" | to_json}}

}

{{ end }}
