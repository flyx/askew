package walker

import (
	"github.com/flyx/net/html"
	"github.com/flyx/net/html/atom"
)

// AskewElements contains all custom elements employed by askew and instructs
// the HTML parser to process them properly.
var AskewElements = []html.CustomElement{
	{Name: "a:component", ProcessLike: atom.Template},
	{Name: "a:macro", ProcessLike: atom.Template},
	{Name: "a:include", DisableFosterParenting: true, ProcessLike: atom.Template},
	{Name: "a:slot", DisableFosterParenting: true, ProcessLike: atom.Template},
	{Name: "a:controller", ProcessLike: atom.Template},
	{Name: "a:handler", ProcessLike: atom.Template},
	{Name: "a:data", ProcessLike: atom.Template},
	{Name: "a:package", ProcessLike: atom.Template},
	{Name: "a:import", ProcessLike: atom.Template},
	{Name: "a:text", ProcessLike: atom.Span},
	{Name: "a:embed", DisableFosterParenting: true, ProcessLike: atom.Template},
	{Name: "a:construct", ProcessLike: atom.Div},
	{Name: "a:site", ProcessLike: atom.Html},
}
