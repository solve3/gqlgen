package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/99designs/gqlgen/_examples/todo"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
)

func main() {
	srv := handler.New(todo.NewExecutableSchema(todo.New()))
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetRecoverFunc(func(ctx context.Context, err any) (userMessage error) {
		// send this panic somewhere
		log.Print(err)
		debug.PrintStack()
		return errors.New("user message on panic")
	})

	http.Handle("/", playground.Handler("Todo", "/query"))
	http.Handle("/query", srv)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
