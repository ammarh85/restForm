package rest

import (
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
)

var jsonMetadata = `{
  "name": "tibco-rest",
  "type": "flogo:activity",
  "ref": "github.com/TIBCOSoftware/flogo-contrib/activity/restForm",
  "version": "0.0.1",
  "title": "Invoke REST Service Form",
  "description": "Simple REST Activity Form",
  "homepage": "https://github.com/TIBCOSoftware/flogo-contrib/tree/master/activity/rest",
  "inputs":[
    {
      "name": "method",
      "type": "string",
      "required": true,
      "allowed" : ["GET", "POST", "PUT", "PATCH", "DELETE"]
    },
    {
      "name": "uri",
      "type": "string",
      "required": true
    },
    {
      "name": "pathParams",
      "type": "params"
    },
    {
      "name": "queryParams",
      "type": "params"
    },
    {
      "name": "content",
      "type": "any"
    },
    {
      "name": "formInput",
      "type": "any"
    }
  ],
  "outputs": [
    {
      "name": "result",
      "type": "any"
    }
  ]
}
`

// init create & register activity
func init() {
	md := activity.NewMetadata(jsonMetadata)
	activity.Register(NewActivity(md))
}
