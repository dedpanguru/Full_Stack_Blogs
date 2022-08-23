package main

import (
	"context"
	"fmt"
	"full_stack_blog/db"
	"full_stack_blog/routes"
	"net/http"
	"os"
)

func main() {
	ctx := context.Background()
	// set up db
	database, err := db.EstablishConnection(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer func() {
		if err = database.Client.Disconnect(database.Context); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}()
	// set up router
	router := routes.NewRouterWIthDB(database)
	// start router
	if err := http.ListenAndServe("0.0.0.0:8080", router); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
