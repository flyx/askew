// +build js,!wasm

package askew

// KeepAlive sends the main thread to sleep if compiled for WASM.
// This is required if your main() entry point would exit; otherwise the
// handlers for DOM events wouldn't be called.
//
// Does nothing when using the GopherJS backend.
func KeepAlive() {
}
