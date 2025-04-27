package actions

import (
	"simplecms/graph" // Import the generated package

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gobuffalo/buffalo"
)

// GraphqlHandler serves the main GraphQL endpoint
func GraphqlHandler() buffalo.Handler {
	// Creates a GraphQL server with the generated schema
	h := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	return func(c buffalo.Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil // Indicate the request was handled
	}
}

// PlaygroundHandler serves the GraphQL Playground UI
func PlaygroundHandler() buffalo.Handler {
	h := playground.Handler("GraphQL Playground", "/query") // Endpoint where Playground sends requests

	return func(c buffalo.Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil // Indicate the request was handled
	}
} 