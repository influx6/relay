package relay

import "html/template"

//this contains template pipe functions that come installed with every TemplateDir generated template map
var DefaultTemplateFunctions = template.FuncMap{
	"iseq": IsEqual,
}

//IsEqual checks equality of values the first two values
func IsEqual(args ...interface{}) bool {
	if len(args) <= 0 {
		return false
	}
	if len(args) == 1 {
		return false
	}
	return args[0] == args[1]
}
