{
   "include" : true,

   {{/* Select from current context; calls EC2 api to describe VPC, then encode the result as JSON */}}
   "Vpc" : {{ describe "/Resources/AWS::EC2::VPC/Vpc" . | jsonEncode }},

   {{/* Load from from ./ using relative path notation. Then split into lines and json encode */}}
   "config" : {{ include "script.tpl" . | lines | jsonEncode }},

   {{/* Load from an URL */}}
   "sample" : {{ include "https://httpbin.org/get" }},

   {{/* Load from URL and then parse as JSON then select an attribute */}}
   "originIp" : "{{ include "https://httpbin.org/get" | jsonDecode | ref "/origin" }}"
}