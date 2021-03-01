title: Conditionals and Loops
date: 2021-01-29
----

# Conditionals and Loops

Askew provides two attributes, `a:if` and `a:loop`, that can be applied on any standard HTML element, and also on `<a:construct>`.

`a:if` takes a value which must be a boolean Go expression.
On component instantiation, this expression is evaluated and the element is removed if it evaluates to `false`.

`a:loop` takes a value with the following syntax:

TODO