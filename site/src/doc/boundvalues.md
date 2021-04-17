title: Bound Values
date: 2021-01-29
----

# Bound Values

Bound values bind DOM values to Go values.
The occur in various places as `<bval>` and have the following syntax:

    <args> ::= <name> ( "," <name> )*
    <bval> ::= <id> "(" [ <args> ] ")"

The `<id>` selects the kind of the bound value.
All availalbe kinds are specified in the following sections.
In `<args>` a list of arguments in the form of names are given.
Their interpretation depends on the bound value's kind.

## `self`

Self takes no arguments, its syntax is always `self()`.
It binds the DOM node of the current element.
Its Go type will be guessed to be `js.Value` and must not be set to anything else.

`self()` is mainly a fallback that allows you to do things that are not possible with Askew otherwise.
For example, you can use it to dynamically add and remove event handlers from a node.
If you want to access or modify properties of the node, you should use a bound value kind that provides direct access to that property instead, if one is available.

## `dataset`

This binds an item in the current element's *dataset* property.
You must give exactly one argument which must be the name of the item in the dataset.

The following example defines two buttons in a component that call the same handler when clicked, but each will provide its own ID as parameter:

```html
<a:component name="Buttons">
  <a:handlers>
    click(id string)
  </a:handlers>
  <button a:capture="click:click(id=dataset(id)"
          data-id="button1">Button 1</button>
  <button a:capture="click:click(id=dataset(id))"
          data-id="button2">Button 2</button>
</a:component>
```

## `prop`

This binds a property of the curent DOM node.
You must give exactly one argument which must be the name of the property.

We already saw an example of this bound value in the previous chapter.

## `style`

This binds a value in the `style` property of the current DOM node.
You must give exactly one argument which must be the name of the style value.

The following example defines a component that renders colored text, where the color is given as parameter:

```html
<a:component name="ColoredText" params="color string">
  <span a:assign="style(color) = color">My text</span>
</a:component>
```

## `class`

This binds a part of the `classList` of the current DOM node.
You must give at least one class name as argument, but can give multiple.
This bound value is used to switch classes on a certain element on and off.

If you give one argument, this bound value will be default map to a **`bool`** value.
Setting it to **`true`** will add the class with the given name to the `classList`, setting it to **`false`** will remove it.

If you give multiple arguments, the value will by default map to an **`int`** value.
Setting it to `0` will remove all named classes from the list, setting it to `1` will add the first class to the list removing all others, and so on.

## `form`

This binds a form element of the current form.
You must give exactly one name as argument, which must be the name of a form element.
This bound value may occur on a `<form>` element or any element that is contained in a `<form>` and always refers to that form.

The form's `elements` DOM property is used to access the element.
If the form element is a radio button group, the value will be default map to a **`string`** value.
Retrieving the value will give you the value of the currently selected radio button, setting it will check the radio button with the given value.

If the form element is not a radio button, the value will also map to a **`string`** by default.
However, now it directly sets and retrieves the `value` property of the linked form element.

## `event`

This bound value may only be used inside `a:capture`.
It gives access to the JavaScript element that has been captured as `js.Value` and cannot use a different type.

## `go`

This bound value may only be used inisde `a:capture`.
With `go(…)`, you can give an arbitrary Go expression in `…` as argument for a handler call.
You can access the current component's object with `o`.
So if you want to give the current value of a data field `foo` as argument to a handler, you would write `go(o.foo)`.