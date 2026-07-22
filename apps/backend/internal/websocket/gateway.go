// Package websocket will host Charli's realtime gateway: one goroutine per
// session, streaming agent steps to the extension and receiving page context.
//
// Deferred to Phase 0 (the chat round-trip), where a websocket library
// (e.g. github.com/coder/websocket) is added and wired into internal/app.
package websocket
