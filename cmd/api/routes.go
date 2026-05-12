package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {

	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	dynamic := alice.New(app.authenticate)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	// Protected (authenticated-only) application routes
	protected := dynamic.Append(app.requireActivatedUser)
	router.Handler(http.MethodGet, "/v1/movies", protected.ThenFunc(app.listMoviesHandler))
	router.Handler(http.MethodPost, "/v1/movies", protected.ThenFunc(app.createMovieHandler))
	router.Handler(http.MethodGet, "/v1/movies/:id", protected.ThenFunc(app.showMovieHandler))
	router.Handler(http.MethodPatch, "/v1/movies/:id", protected.ThenFunc(app.updateMovieHandler))
	router.Handler(http.MethodDelete, "/v1/movies/:id", protected.ThenFunc(app.deleteMovieHandler))

	standard := alice.New(app.recoverPanic, app.logRequest, app.rateLimit)

	return standard.Then(router)
}
