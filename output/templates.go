package output

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/flyx/askew/data"
)

var fileHeader = template.Must(template.New("fileHeader").Funcs(template.FuncMap{
	"FormatImport": func(alias, path string) string {
		if filepath.Base(path) == alias {
			return "\"" + path + "\""
		}
		return alias + " \"" + path + "\""
	},
}).Parse(`
package {{.PackageName}}

import (
	"syscall/js"
	"github.com/flyx/askew/runtime"
	{{- range $alias, $path := .Imports }}
	{{FormatImport $alias $path}}{{ end }}
)
`))

var file = template.Must(template.New("file").Funcs(template.FuncMap{
	"Wrapper":      wrapperForType,
	"PathItems":    pathItems,
	"NameForBound": nameForBound,
	"Last":         last,
	"TWrapper": func(t *data.ParamType, name string) string {
		return wrapperForType(*t) + "{BoundValue: " + name + "}"
	},
	"IsBool": func(t *data.ParamType) bool {
		return t != nil && t.Kind == data.BoolType
	},
	"GenParams": func(params []data.Param) string {
		var items []string
		for _, p := range params {
			items = append(items, p.String())
		}
		return strings.Join(items, ", ")
	},
	"GenReturns": func(value *data.ParamType) string {
		if value == nil {
			return ""
		}
		return value.String()
	},
	"GenArgs": func(params []data.BoundParam) string {
		items := make([]string, 0, len(params))
		for _, p := range params {
			items = append(items, fmt.Sprintf("&p%s", p.Param))
		}
		return strings.Join(items, ", ")
	},
	"ClassNames": func(list []string) string {
		var b strings.Builder
		first := true
		for _, item := range list {
			if first {
				first = false
			} else {
				b.WriteString(", ")
			}
			b.WriteByte('"')
			b.WriteString(item)
			b.WriteByte('"')
		}
		return b.String()
	},
	"IsFormValue": func(bk data.BoundKind) bool {
		return bk == data.BoundFormValue
	},
	"IsClassValue": func(bk data.BoundKind) bool {
		return bk == data.BoundClass
	},
	"IsEventValue": func(bk data.BoundKind) bool {
		return bk == data.BoundEventValue
	},
	"IsSelfValue": func(bk data.BoundKind) bool {
		return bk == data.BoundSelf
	},
	"NeedsSelf": func(params []data.BoundParam) bool {
		for _, p := range params {
			if p.Value.Kind != data.BoundEventValue {
				return true
			}
		}
		return false
	},
	"TypeForKind": func(bk data.BoundKind) string {
		switch bk {
		case data.BoundProperty:
			return "BoundProperty"
		case data.BoundStyle:
			return "BoundStyle"
		case data.BoundData:
			return "BoundData"
		case data.BoundClass:
			return "BoundClasses"
		case data.BoundSelf:
			return "BoundSelf"
		default:
			panic("unknown BoundKind")
		}
	},
	"GenCallParams": func(params []data.Param) string {
		items := make([]string, 0, len(params))
		for _, p := range params {
			items = append(items, p.Name+" runtime.BoundValue")
		}
		return strings.Join(items, ", ")
	},
	"GenTypedArgs": func(params []data.Param) string {
		items := make([]string, 0, len(params))
		for _, p := range params {
			items = append(items, fmt.Sprintf("_%s.Get()", p.Name))
		}
		return strings.Join(items, ", ")
	},
	"GenComponentParams": func(params []data.ComponentParam) string {
		items := make([]string, 0, len(params))
		for _, p := range params {
			items = append(items, fmt.Sprintf("%s %s", p.Name, p.Type))
		}
		return strings.Join(items, ", ")
	},
	"ListParamVars": func(params []data.ComponentParam) string {
		items := make([]string, 0, len(params))
		for _, p := range params {
			items = append(items, p.Name)
		}
		return strings.Join(items, ", ")
	},
	"FieldType": func(e data.Embed) string {
		if e.T == "" {
			switch e.Kind {
			case data.OptionalEmbed:
				return "runtime.GenericOptional"
			case data.ListEmbed:
				return "runtime.GenericList"
			default:
				panic("unexpected field type")
			}
		}
		var b strings.Builder
		if e.Kind == data.DirectEmbed {
			b.WriteRune('*')
		}
		if e.Ns != "" {
			b.WriteString(e.Ns)
			b.WriteRune('.')
		}
		if e.Kind == data.OptionalEmbed {
			b.WriteString("Optional")
		}
		b.WriteString(e.T)
		if e.Kind == data.ListEmbed {
			b.WriteString("List")
		}
		return b.String()
	},
	"BlockNotEmpty": func(b data.Block) bool {
		return len(b.Assignments) > 0 || len(b.Controlled) > 0
	},
}).Option("missingkey=error").Parse(`
{{- define "Block"}}
  {{- range .Assignments}}
	{
		{{- if IsFormValue .Target.Kind}}
		var tmp runtime.BoundFormValue
		tmp.Init(runtime.WalkPath(block, {{PathItems .Path .Target.FormDepth}}), "{{.Target.ID}}", {{.Target.IsRadio}})
		{{- else if IsClassValue .Target.Kind}}
		var tmp runtime.BoundClasses
		tmp.Init(runtime.WalkPath(block, {{PathItems .Path .Target.FormDepth}}), []string{ {{ClassNames .Target.IDs}} })
		{{- else if IsSelfValue .Target.Kind}}
		var tmp runtime.BoundSelf
		tmp.Init(runtime.WalkPath(block, {{PathItems .Path 0}}))
		{{- else}}
		var tmp runtime.{{TypeForKind .Target.Kind}}
		tmp.Init(runtime.WalkPath(block, {{PathItems .Path 0}}), "{{.Target.ID}}")
		{{- end}}
		runtime.Assign(&tmp, {{.Expression}})
	}
	{{- end}}

	{{- range .Controlled}}
	{{- if eq .Kind 0}}
	if {{.Expression}} {
		{{if BlockNotEmpty .Block}}
		block := runtime.WalkPath(block, {{PathItems .Path 0}})
		{{template "Block" .Block}}
		{{- end}}
	} else {
		_item := runtime.WalkPath(block, {{PathItems .Path 0}})
		_parent := _item.Get("parentNode")
		_parent.Call("replaceChild", js.Global().Get("document").Call("createComment", "removed"), _item)
	}
	{{- else }}
	{
		_orig := runtime.WalkPath(block, {{PathItems .Path 0}})
		_parent := _orig.Get("parentNode")
		_next := _orig.Get("nextSibling")
		_parent.Call("removeChild", _orig)
		for {{.Index}}{{with .Variable}}, {{.}}{{end}} := range {{.Expression}} {
			block := _orig.Call("cloneNode", true)
			{{template "Block" .Block}}
			_parent.Call("insertBefore", block, _next)
		}
	}
	{{- end}}
	{{- end}}
{{- end}}

{{- range .Components}}
{{- if .Controller}}
// {{.Name}}Controller can be implemented to handle external events
// generated by {{.Name}}
type {{.Name}}Controller interface {
	{{- range $name, $handler := .Controller }}
	{{$name}}({{GenParams $handler.Params }}){{GenReturns $handler.Returns}}
	{{- end }}
}
{{- end}}

// {{.Name}} is a DOM component autogenerated by Askew
type {{.Name}} struct {
	αcd runtime.ComponentData
	{{- if .Controller }}
	// Controller is the adapter for events generated from this component.
	// if nil, events that would be passed to the controller will not be handled.
	Controller {{.Name}}Controller
	{{- end}}
	{{- range .Variables }}
	{{.Variable.Name}} {{Wrapper .Variable.Type}}
	{{- end}}
	{{- range .Fields}}
	{{.Name}} {{.Type}}
	{{- end}}
	{{- range .Embeds }}
	{{.Field}} {{FieldType .}}
	{{- end}}
}

// {{.ConstructorName}} creates a new component and initializes it with Init.
func {{.ConstructorName}}({{GenComponentParams .Parameters}}) *{{.Name}} {
	ret := new({{.Name}})
	ret.Init({{ListParamVars .Parameters}})
	return ret
}

// Data returns the object containing the component's DOM nodes.
// It implements the runtime.Component interface.
func (o *{{.Name}}) Data() *runtime.ComponentData {
	return &o.αcd
}

// Init initializes the component, discarding all previous information.
// The component is initially a DocumentFragment until it gets inserted into
// the main document. It can be manipulated both before and after insertion.
func (o *{{.Name}}) Init({{GenComponentParams .Parameters}}) {
	o.αcd.Init(runtime.InstantiateTemplateByID("{{.ID}}"))
	{{ range .Fields }}
	{{- if .DefaultValue }}o.{{.Name}} = {{.DefaultValue}}
	{{end}}
	{{- end}}
	{{- range .Variables }}
	{{- if IsFormValue .Value.Kind}}
	o.{{.Variable.Name}}.BoundValue = runtime.NewBoundFormValue(&o.αcd, "{{.Value.ID}}", {{.Value.IsRadio}}, {{PathItems .Path .Value.FormDepth}})
	{{- else if IsClassValue .Value.Kind}}
	o.{{.Variable.Name}}.BoundValue = runtime.NewBoundClasses(&o.αcd, []string{ {{ClassNames .Value.IDs}} }, {{PathItems .Path 0}})
	{{- else if IsSelfValue .Value.Kind}}
	o.{{.Variable.Name}}.BoundValue = runtime.NewBoundSelf(&o.αcd, {{PathItems .Path 0}})
	{{- else}}
	o.{{.Variable.Name}}.BoundValue = runtime.New{{TypeForKind .Value.Kind}}(&o.αcd, "{{.Value.ID}}", {{PathItems .Path 0}})
	{{- end}}
	{{- end}}
	{{- if BlockNotEmpty .Block}}
	{
		block := o.αcd.Walk()
		{{- template "Block" .Block}}
	}
	{{- end}}
	{{- range .Captures}}
	{
		src := o.αcd.Walk({{PathItems .Path 0}})
		{{- range .Mappings}}
		{
			wrapper := js.FuncOf(func(this js.Value, arguments []js.Value) interface{} {
				{{- if NeedsSelf .ParamMappings}}
				self := arguments[0].Get("currentTarget")
				{{- end}}
				{{- range .ParamMappings}}
				var p{{.Param}} runtime.{{NameForBound .Value.Kind}}
				{{- if IsFormValue .Value.Kind}}
				p{{.Param}}.Init(self.Call("closest", "form"), "{{.Value.ID}}", {{.Value.IsRadio}})
				{{- else if IsEventValue .Value.Kind}}
				p{{.Param}}.Init(arguments[0], "{{.Value.ID}}")
				{{- else}}
				p{{.Param}}.Init(self, "{{.Value.ID}}")
				{{- end}}
				{{- end}}
				{{- if eq .Handling 0}}
				o.αcall{{.Handler}}({{GenArgs .ParamMappings}})
				arguments[0].Call("preventDefault")
				{{- else if eq .Handling 2}}
				if o.αcall{{.Handler}}({{GenArgs .ParamMappings}}) {
					arguments[0].Call("preventDefault")
				}
				{{- else }}
				o.αcall{{.Handler}}({{GenArgs .ParamMappings}})
				{{- end}}
				return nil
			})
			src.Call("addEventListener", "{{.Event}}", wrapper)
		}
		{{- end}}
	}
	{{- end}}
	{{- range .Embeds }}
	{
		container := o.αcd.Walk({{PathItems .Path 1}})
		{{- if eq .Kind 0}}
		o.{{.Field}} = {{with .Ns}}{{.}}.{{end}}{{.ConstructorName}}({{.Args.Raw}})
		o.{{.Field}}.InsertInto(container, container.Get("childNodes").Index({{Last .Path}}))
		{{- if .Control}}
		o.{{.Field}}.Controller = o
		{{- end}}
		{{- else}}
		o.{{.Field}}.Init(container, {{Last .Path}})
		{{- if .Control}}
		o.{{.Field}}.DefaultController = o
		{{- end}}
		{{$e := .}}
		{{- range .ConstructorCalls}}
		{{- if eq .Kind 1}}
		if {{.Expression}} {
		{{- else if eq .Kind 2}}
		for {{.Index}}, {{.Variable}} := range {{.Expression}} {
		{{- end}}
		{{- if eq $e.Kind 2}}
		o.{{$e.Field}}.Set(
		{{- else}}
		o.{{$e.Field}}.Append(
		{{- end}}{{with $e.Ns}}{{.}}.{{end}}{{$e.ConstructorName}}({{.Args.Raw}}))
		{{- if ne .Kind 0}}
		}
		{{- end}}
		{{- end}}
		{{- end}}
	}
	{{- end}}
	{{- if .Init}}
	o.init({{ListParamVars .Parameters}})
	{{- end}}
}

// InsertInto inserts this component into the given object.
// The component will be in inserted state afterwards.
//
// The component will be inserted in front of 'before', or at the end if 'before' is 'js.Undefined()'.
func (o *{{.Name}}) InsertInto(parent js.Value, before js.Value) {
	o.αcd.DoInsert(parent, before)
	{{- range .Embeds}}
	{{- if ne .Kind 0}}
	{{- if .T}}
	o.{{.Field}}.αmgr.UpdateParent(o.αcd.DocumentFragment(), parent, before)
	{{- else}}
	o.{{.Field}}.DoUpdateParent(o.αcd.DocumentFragment(), parent, before)
	{{- end}}
	{{- end}}
	{{- end}}
	{{- if .OnInclude}}
	o.onInclude()
	{{- end}}
}

// Extract removes this component from its current parent.
// The component will be in initial state afterwards.
func (o *{{.Name}}) Extract() {
	o.αcd.DoExtract()
	{{- range .Embeds}}
	{{- if ne .Kind 0}}
	{{- if .T}}
	o.{{.Field}}.αmgr.UpdateParent(o.αcd.First().Get("parentNode"), o.αcd.DocumentFragment(), js.Undefined())
	{{- else}}
	o.{{.Field}}.DoUpdateParent(o.αcd.First().Get("parentNode"), o.αcd.DocumentFragment(), js.Undefined())
	{{- end}}
	{{- end}}
	{{- end}}
	{{- if .OnExclude}}
	o.onExclude()
	{{- end}}
}

// Destroy destroys this element (and all contained components). If it is
// currently inserted anywhere, it gets removed before.
func (o *{{.Name}}) Destroy() {
	{{- range .Embeds}}
	{{- if eq .Kind 0}}
	o.{{.Field}}.Destroy()
	{{- else if eq .Kind 1}}
	o.{{.Field}}.DestroyAll()
	{{- else}}
	o.{{.Field}}.Set(nil)
	{{- end}}
	{{- end}}
	o.αcd.DoDestroy()
}

{{$cName := .Name}}
{{- range $hName, $h := .Handlers}}
func (o *{{$cName}}) αcall{{$hName}}({{GenCallParams $h.Params}}) {{if IsBool $h.Returns}}bool{{end}} {
	{{- range $h.Params}}
	_{{.Name}} := {{TWrapper .Type .Name}}
	{{- end}}
	{{if IsBool $h.Returns}}return {{end}}o.{{$hName}}({{GenTypedArgs $h.Params}})
}
{{- end}}
{{- range $hName, $m := .Controller}}
{{- if $m.CanCapture}}
func (o *{{$cName}}) αcall{{$hName}}({{GenCallParams $m.Params}}) {{if IsBool $m.Returns}}bool{{end}} {
	if o.Controller == nil {
		return{{if IsBool $m.Returns}} false{{end}}
	}
	{{- range $m.Params}}
	_{{.Name}} := {{TWrapper .Type .Name}}
	{{- end}}
	{{if IsBool $m.Returns}}return {{end}}o.Controller.{{$hName}}({{GenTypedArgs $m.Params}})
}
{{- end}}
{{- end}}

{{if .NeedsList}}
// {{.Name}}List is a list of {{.Name}} whose manipulation methods auto-update
// the corresponding nodes in the document.
type {{.Name}}List struct {
	αmgr runtime.ListManager
	αitems []*{{.Name}}
	{{- if .Controller}}
	DefaultController {{.Name}}Controller
	{{- end}}
}

// Init initializes the list, discarding previous data.
// The list's items will be placed in the given container, starting at the
// given index.
func (l *{{.Name}}List) Init(container js.Value, index int) {
	l.αmgr = runtime.CreateListManager(container, index)
	l.αitems = nil
}

// Len returns the number of items in the list.
func (l *{{.Name}}List) Len() int {
	return len(l.αitems)
}

// Item returns the item at the current index.
func (l *{{.Name}}List) Item(index int) *{{.Name}} {
	return l.αitems[index]
}

// Append appends the given item to the list.
func (l *{{.Name}}List) Append(item *{{.Name}}) {
	if item == nil {
		panic("cannot append nil to list")
	}
	l.αmgr.Append(item)
	l.αitems = append(l.αitems, item)
	{{- if .Controller}}
	item.Controller = l.DefaultController
	{{- end}}
	return
}

// Insert inserts the given item at the given index into the list.
func (l *{{.Name}}List) Insert(index int, item *{{.Name}}) {
	var prev js.Value
	if index < len(l.αitems) {
		prev = l.αitems[index].αcd.First()
	}
	if item == nil {
		panic("cannot insert nil into list")
	}
	l.αmgr.Insert(item, prev)
	l.αitems = append(l.αitems, nil)
	copy(l.αitems[index+1:], l.αitems[index:])
	l.αitems[index] = item
	{{- if .Controller}}
	item.Controller = l.DefaultController
	{{- end}}
	return
}

// Remove removes the item at the given index from the list and returns it.
func (l *{{.Name}}List) Remove(index int) *{{.Name}} {
	item := l.αitems[index]
	item.Extract()
	copy(l.αitems[index:], l.αitems[index+1:])
	l.αitems = l.αitems[:len(l.αitems)-1]
	return item
}

// Destroy destroys the item at the given index and removes it from the list.
func (l *{{.Name}}List) Destroy(index int) {
	item := l.αitems[index]
	item.Destroy()
	copy(l.αitems[index:], l.αitems[index+1:])
	l.αitems = l.αitems[:len(l.αitems)-1]
}

// DestroyAll destroys all items in the list and empties it.
func (l *{{.Name}}List) DestroyAll() {
	for _, item := range l.αitems {
		item.Destroy()
	}
	l.αitems = l.αitems[:0]
}
{{end}}

{{- if .NeedsOptional}}
// Optional{{.Name}} is a nillable embeddable container for {{.Name}}.
type Optional{{.Name}} struct {
	αcur *{{.Name}}
	αmgr runtime.ListManager
	{{- if .Controller}}
	DefaultController {{.Name}}Controller
	{{- end}}
}

// Init initializes the container to be empty.
// The contained item, if any, will be placed in the given container at the
// given index.
func (o *Optional{{.Name}}) Init(container js.Value, index int) {
	o.αmgr = runtime.CreateListManager(container, index)
	o.αcur = nil
}

// Item returns the current item, or nil if no item is assigned
func (o *Optional{{.Name}}) Item() *{{.Name}} {
	return o.αcur
}

// Set sets the contained item destroying the current one.
// Give nil as value to simply destroy the current item.
func (o *Optional{{.Name}}) Set(value *{{.Name}}) {
	if o.αcur != nil {
		o.αcur.Destroy()
	}
	o.αcur = value
	if value != nil {
		o.αmgr.Append(value)
		{{- if .Controller}}
		value.Controller = o.DefaultController
		{{- end}}
	}
}

// Remove removes the current item and returns it.
// Returns nil if there is no current item.
func (o *Optional{{.Name}}) Remove() runtime.Component {
	if o.αcur != nil {
		ret := o.αcur
		ret.Extract()
		o.αcur = nil
		return ret
	}
	return nil
}

{{- end}}
{{- end}}
`))

var skeleton = template.Must(template.New("skeleton").Funcs(template.FuncMap{
	"PathItems": pathItems,
	"Last":      last,
}).Parse(`
{{if .VarName}}
// {{.VarName}} holds the embedded components of the document's skeleton
var {{.VarName}} = struct {
	{{- range .Embeds}}
		// {{.Field}} is part of the main document.
		{{- if eq .Kind 0}}
			{{.Field}} *{{with .Ns}}{{.}}.{{end}}{{.T}}
		{{- else if eq .Kind 1}}
			{{- if .T}}
				{{.Field}} {{with .Ns}}{{.}}.{{end}}{{.T}}List
			{{- else}}
				{{.Field}} runtime.GenericList
			{{- end}}
		{{- else}}
			{{- if .T}}
				{{.Field}} {{with .Ns}}{{.}}.{{end}}Optional{{.T}}
			{{- else}}
				{{.Field}} runtime.GenericOptional
			{{- end}}
		{{- end}}
	{{- end -}}
}{
	{{- range .Embeds}}
		{{- if eq .Kind 0}}
			{{.Field}}: {{with .Ns}}{{.}}.{{end}}New{{.ConstructorName}}({{.Args.Raw}}),
		{{- else if eq .Kind 1}}
			{{- if .T}}
				{{.Field}}: {{with .Ns}}{{.}}.{{end}}{{.T}}List{},
			{{- else}}
				{{.Field}}: runtime.GenericList{},
			{{- end}}
		{{- else}}
			{{- if .T}}
				{{.Field}}: {{with .Ns}}{{.}}.{{end}}Optional{{.T}}{},
			{{- else}}
				{{.Field}}: runtime.GenericOptional{},
			{{- end}}
		{{- end}}
	{{- end}}
}
{{- else}}
	{{range .Embeds}}
		// {{.Field}} is part of the main document.
		{{- if eq .Kind 0}}
			var {{.Field}} = {{with .Ns}}{{.}}.{{end}}{{.ConstructorName}}({{.Args.Raw}})
		{{- else if eq .Kind 1}}
			{{- if .T}}
				var {{.Field}} {{with .Ns}}{{.}}.{{end}}{{.T}}List
			{{- else}}
				var {{.Field}} runtime.GenericList
			{{- end}}
		{{- else}}
			{{- if .T}}
				var {{.Field}} {{with .Ns}}{{.}}.{{end}}Optional{{.T}}
			{{- else}}
				var {{.Field}} runtime.GenericOptional
			{{- end}}
		{{- end}}
	{{- end}}
{{- end}}

{{$varName := .VarName}}
func init() {
	html := js.Global().Get("document").Get("childNodes").Index(1)
	{{- range .Embeds}}
	{{- if eq .Kind 0}}
	{
		container := runtime.WalkPath(html, {{PathItems .Path 1}})
		{{with $varName}}{{.}}.{{end}}{{.Field}}.InsertInto(container, container.Get("childNodes").Index({{Last .Path}}))
	}
	{{- else}}
	{{with $varName}}{{.}}.{{end}}{{.Field}}.Init(runtime.WalkPath(html, {{PathItems .Path 1}}), {{Last .Path}})
	{{- end}}
	{{- end}}
}
`))
