package main

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type rules struct {
	app.Compo
}

func newRules() *rules {
	return &rules{}
}

func (r *rules) Render() app.UI {
	return app.Div().ID("rules").Body(
		app.A().Href("#").Text("Rules | Terms").OnClick(r.openPopup),
		app.Div().ID("overlay").OnClick(r.closePopup),
		app.Div().ID("popup").Body(
			app.Div().Class("list-container").Body(
				app.Div().ID("rules").Body(
					app.H2().Text("Rules"),
					app.Ul().Body(
						app.Ol().Body(
							app.Li().Text("Cyber-derive is meant to be used for inner-city deliveries only."),
							app.Li().Text("Requesters request a delivery to be made by posting coordinates of pick up and drop off location as well as entry time for competition and delivery completion."),
							app.Li().Text("In case of a single courier participant he/she is automatically a winner."),
							app.Li().Text("In case of competing couriers an algorithm decides based on 3 factors - energy, safety and speed."),
							app.Li().Text("These 3 stats are gained via magic cards which can be equipped before each game."),
							app.Li().Text("After all avatar slots are filled a player can swap only one card per competition."),
							app.Li().Text("After a competition is won the winner gets to complete the delivery."),
							app.Li().Text("After a delivery is confirmed by the requester the courier gets a reward in the form of bucket of special cards to choose from."),
							app.Li().Text("He/she can choose to swap only one card per win."),
							app.Li().Text("In case of not delivered parcel the requester can file a report."),
							app.Li().Text("In case of theft or abuse reported the courier gets banned for good from the platform automatically."),
							app.Li().Text("In case of non-presence the couriers gets a warning, on a second one a ban is automatically applied."),
							app.Li().Text("Packages limit - up to 5 kg"),
							app.Li().Text("Dimensions of items - have to fit in a regular backpack."),
						),
					),
				),
				app.Div().ID("terms").Body(
					app.H2().Text("Terms"),
					app.Ul().Body(
						app.Ol().Body(
							app.Li().Text("Cyber-derive is a non-commercial self-managed gamified delivery p2p app."),
							app.Li().Text("There is no personal data collected."),
							app.Li().Text("The only trace of participants is their IP which is associated with the unique Peer ID used in the system."),
							app.Li().Text("Cyber-derive doesn't collect information about item contents or participants."),
							app.Li().Text("In case of abuse/theft/fraud IP of courier can be found and reported to public authorities via Peer ID."),
							app.Li().Text("As a non-commercial p2p app there is no moderation, censorship or administration as the app is co-hosted on all participant devices and no one can take it down singlehandedly."),
							app.Li().Text("In case of theft/abuse/fraud/accident Stateless Minds is non-liable to any damages or accusations."),
							app.Li().Text("In case of illegal items Stateless Minds is non-liable for the content."),
							app.Li().Text("Stateless Minds can not be held responsible for the operation and data in the game."),
							app.Li().Text("As an open-source game licensed under MIT Cyber-derive can be reproduced and forked by anyone."),
						),
					),
				),
			),
		),
	)
}

func (r *rules) openPopup(ctx app.Context, e app.Event) {
	app.Window().GetElementByID("popup").Get("style").Set("display", "block")
	app.Window().GetElementByID("overlay").Get("style").Set("display", "block")
}

func (r *rules) closePopup(ctx app.Context, e app.Event) {
	app.Window().GetElementByID("popup").Get("style").Set("display", "none")
	app.Window().GetElementByID("overlay").Get("style").Set("display", "none")
}
