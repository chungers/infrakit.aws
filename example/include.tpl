{
   "include" : true,

   {{/* Select from current context; calls EC2 api to describe VPC, then encode the result as JSON */}}
   "Vpc" : {{ describe "/Resources/AWS::EC2::VPC/Vpc" . | jsonEncode }},

   {{/* Load from from ./ using relative path notation. Then split into lines and json encode */}}
   "config" : {{ include "script.tpl" . | lines | jsonEncode }},

   {{/* Load from an URL */}}
   "sample" : {{ include "https://httpbin.org/get" }},

   {{/* Load from URL and then parse as JSON then select an attribute */}}
   "originIp" : "{{ include "https://httpbin.org/get" | jsonDecode | ref "/origin" }}",

   {{/* Load from unix socket -- be sure to export SOCKET_DIR=dir and hostname is the filename */}}
   "infrakitInstanceFile" : {{ include "unix://instance-file/meta/api.json" }},

   {{/*
   Select from array with selector expression matching field value.  Note the selector can be itself complex and
   descends into each element of the array to select the value for comparison.  The node that has the first match
   is then used to continue past the [] of the selection expression.7
   */}}
   {{ with $pluginMeta := "unix://instance-file/meta/api.json" }}
   {{ with $select := "Interfaces[Name=Instance]/Methods['Request/method'=Instance.Validate]/Request/params" }}
   "infrakitInstanceFile_ProvisionProperties" : {{ include $pluginMeta |jsonDecode | ref $select | jsonEncode }}
   {{ end }}
   {{ end }}
}