{
   "include" : true,
   "Vpc" : {{ describe "/Resources/AWS::EC2::VPC/Vpc" . | jsonMarshal }},
   "config" : {{ include "script.tpl" . | lines | jsonMarshal }}
}