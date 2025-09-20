package main

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type notFound struct {
	app.Compo
}

func (n *notFound) Render() app.UI {
	return app.Div().Class("full-screen").Body(
		app.H1().Class("not-found").Text(404),
	)
}
