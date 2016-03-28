package frederic

import (
	"text/template"
)
// put this in a separate file as I couldn't figure out how to declare
// an html/template and a text/template in the same file

var txttemplates = template.Must(template.New("csv").ParseGlob("*.csv"))

