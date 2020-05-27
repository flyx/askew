# Template Based Components (TBC)

TBC is a tool that auto-generates Go code for managing HTML components in GopherJS.
You write HTML `<template>` nodes including some meta-tags and -elements, and TBC autogenerates Go types from each one that can be used to instantiate the template.
The Go types let you access and alter dynamic content in the instantiated HTML subtree.
You can also capture and handle node events.

TBC is an unopinionated library and doesn't care how your HTML looks, how you communicate with a server, what database you use etc.
Its sole purpose is to relieve you from the burden of having to deal with the DOM API.

TBC is written for browsers that support the `<template>` element.
If yours doesn't, tough luck.

## Usage

The pipeline goes like this:

<img src="./pipeline.svg">

The input file must contain only `<template>` and `<tbc:marco>` elements at top level.

Each template will generate a `<name>.go` file containing a `struct` type generated from the template.
This type will have value bindings for all items in the HTML you declared to be dynamic.
A value binding provides `Set` and `Get` methods.

You need to implement methods for each event handler you have declared in the template on the generated type.
Go allows you to do this in a different file since the generated file will be overridden on changes in the input templates.

TBD: Syntax definition, examples.

## License

MIT