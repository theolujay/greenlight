package main

import (
	"fmt"
	"net/http"
)

// The recoverPanic() method is a middleware for the server to send a
// 500 Internal Server Error when it panics rather than just closing
// the HTTP connection with no context. It doesn't recover panics in
// other goroutines, so ensure to recover panics from within those
// goroutines.
//
// NOTE:
//
// http.Server may still automatically generate and send plain-text
// HTTP responses in the following scenarios:
//
// - The HTTP request specifies an unsupported HTTP protocol version.
//
// - The HTTP request contains a missing or invalid Host header, or multiple Host headers.
//
// - The HTTP request contains a empty Content-Length header.
//
// - The HTTP request contains an unsupported Transfer-Encoding header.
//
// - The size of the HTTP request headers exceeds the server’s MaxHeaderBytes setting.
//
// - The client makes a HTTP request to a HTTPS server.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
