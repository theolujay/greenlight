package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/theolujay/greenlight/internal/data"
	"github.com/theolujay/greenlight/internal/validator"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.logger.Info(
			fmt.Sprintf(
				"%s - %s %s %s",
				r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI()),
		)
		next.ServeHTTP(w, r)
	})
}

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

// The rateLimit() method is a middleware that implements rate-limiter
// based on the IP address of the client.
func (app *application) rateLimit(next http.Handler) http.Handler {
	// Hold the rate limiter and last seen time for each client.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	// A mutex and a map to hold the clients' IP addresses
	// and rate limiters.
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Launch a background goroutine which removes old entries from
	// the clients map once every minute.
	go func() {
		for {
			time.Sleep(time.Minute)

			// Lock the mutex to prevent any rate limiter checks from
			// happening while the cleanup is taking place.
			mu.Lock()

			// Loop through all clients. If they haven't been seen
			// within the last three minutes, delete the corresponding
			// entry from the map.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			// Unlock the mutext when the clean up is complete
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			// Get client's IP address
			ip := realip.FromRequest(r)

			mu.Lock()

			// Check to see if the IP address already exists in the map. If
			// it doesn't , then initialize the new rate limiter and add the
			// IP address and limier to the map.
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(
						rate.Limit(app.config.limiter.rps),
						app.config.limiter.burst,
					),
				}
			}

			// Update the last seen time for the client
			clients[ip].lastSeen = time.Now()

			// Call the Allow() method on the rate limiter for the current
			// IP address. If the request isn't allowed, unlock the mutex
			// and send a 429 Too Many Requests response.
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			// Make sure to unlock the mutex before calling the next handler
			// in the chain. And notice how `defer` isn't used to unlock the
			// mutext, as that would mean that the mutext isn't unlocked
			// until all the handlers downstream of this middleware have
			// also returned.
			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the "Vary: Authorization" header to the response. This indicates to any
		// caches that the response may vary based on the value of the Authorization
		// header in the request. Without it, a cache might serve the same response
		// to different users, which is a serious security bug.
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		v := validator.New()

		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		r = app.contextSetUser(r, user)

		next.ServeHTTP(w, r)

	})
}

func (app *application) requireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireActivatedUser(next http.Handler) http.Handler {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	return app.requireAuthenticatedUser(fn)
}
