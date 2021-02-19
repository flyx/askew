title: Debugging
date: 2021-01-10
----

# Debugging

Compiling with GopherJS will give you a source map together with the generated JavaScript file.
However, this source map is only useful if your web server serves the referenced Go source files.

Of course, the source code may span several modules, so it is not as simple as putting the sources of your current module onto the web server.
However, Go offers a simple solution:

    go mod vendor

This will create a directory `vendor` with all sources of all modules used by your module, except for the standard library.
If you want to debug into the standard library, you will need to copy over `$GOROOT/libexec/src`.

All the `*.go` files in the `vendor` directory will then need to be served by your web server according to their directory hierarchy, rooting at `vendor`.
How you set this up depends on your web server and is out of scope for this documentation.

Be aware that the presence of the `vendor` directory indicates to the `go` compiler that you want to use Vendoring, which will make it use the local sources within that directory instead of the original modules.
You'll want to rename or move the directory somewhere else.

If everything works correctly, you should be able to access your Go sources via your browser's development tools.
You can set breakpoints there and step through your Go statements, though the debugger will throw you back to the generated JavaScript for lines that do not correspond to a Go source line.