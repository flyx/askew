title: Macros
date: 2021-01-28
----

# Macros

Sometimes you replicate parts of your HTML structure in different components.
Askew's *macros* are a tool to avoid duplicate code by placing it in a macro and including that macro in different places.

Inclusions of macros are replaced with the macro content before any other processing takes places.
You can include macros from a different package, but you can only include macros defined in the same module.
The reason for this is that Askew needs access to the macro's source, which it doesn't have if the macro is defined in a different module.

Macros don't produce any Go code.

## Defining Macros

A macro is defined with `<a:macro>`.
The element requires an attribute `name` which is the identifier of the macro and must be unique for all macros in the same package.
The content of a macro is an HTML subtree, Askew elements and attributes may be used in it.
Names referred to in Askew elements, for example a component parameter name, are only resolved after macro inclusion took place, so you can refer parameters the macro can't see, as long as they exist at the place the macro is included:

```html
<a:macro name="myMacro">
	<!-- handler submit() unknown, must be available where
	     macro is included. -->
	<form a:capture="submit:submit()">
	  <!-- variable labelText also must be available on inclusion -->
		<label><a:text expr="labelText"></a:text></label>
	</form>
</a:macro>
```

A macro can contain one ore more `<a:slot>` elements.
Such elements also require a `name` attribute which is their identifier and must be unique inside the macro.
An `<a:slot>` will be replaced by content defined on the inclusion site, so it acts like a parameter.
The `<a:slot>` element may contain content which will be its default value if no other content is given at the inclusion site:

```html
<a:macro name="hello">
  Hello, <a:slot name="who">World!</a:slot>
</a:macro>
```

## Including Macros

Macros are included via `<a:include>`.
This element must have a name which identifies the macro to include.
The name may contain a package name if the macro is defined in a different package.

`<a:include>` may contain elements, where each element must have an attribute `a:slot` whose value is the name of a slot of the target macro.
That slot will be replaced by this element on inclusion:

```html
<a:include name="hello">
  <span a:slot="who">Karl Koch!</a:slot>
</a:include>
```

Currently, only single elements can be given as a replacement for a slot.
