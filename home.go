package main

import (
	"encoding/json"
	"slices"
	"strings"
	"time"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	shell "github.com/stateless-minds/go-ipfs-api"
)

// home is the main component that displays the home menu. A component is a
// customizable, independent, and reusable UI elemenh. It is created by
// embedding app.Compo into a struch.
type home struct {
	app.Compo
	notificationPermission app.NotificationPermission
	sh                     *shell.Shell
	myPeerID               string
	deliveryJSON           string
	deliveries             []Delivery
}

func (h *home) OnMount(ctx app.Context) {
	h.notificationPermission = ctx.Notifications().Permission()
	switch h.notificationPermission {
	case app.NotificationDefault:
		h.notificationPermission = ctx.Notifications().RequestPermission()
	case app.NotificationDenied:
		app.Window().Call("alert", "In order to use Cyber Dérive notifications needs to be enabled")
		return
	}

	sh := shell.NewShell("localhost:5001")
	h.sh = sh

	myPeer, err := h.sh.ID()
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	// remove c1 after testing
	h.myPeerID = myPeer.ID + "c1"

	// err = h.sh.OrbitDocsDelete(dbDelivery, "all")
	// if err != nil {
	// 	ctx.Notifications().New(app.Notification{
	// 		Title: "Error",
	// 		Body:  err.Error(),
	// 	})
	// }

	// return

	// ctx.Async(func() {
	// 	err := h.sh.OrbitDocsDelete(dbAvatar, "all")
	// 	if err != nil {
	// 		ctx.Notifications().New(app.Notification{
	// 			Title: "Error",
	// 			Body:  err.Error(),
	// 		})
	// 	}
	// })

	// return

	// err = h.sh.OrbitDocsDelete(dbReport, "all")
	// if err != nil {
	// 	ctx.Notifications().New(app.Notification{
	// 		Title: "Error",
	// 		Body:  err.Error(),
	// 	})
	// }

	// return

	ctx.SetState("page", "active")

	h.checkReports(ctx)
}

func (h *home) checkReports(ctx app.Context) {
	ctx.Async(func() {
		reportsJSON, err := h.sh.OrbitDocsGet(dbReport, h.myPeerID)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		if strings.TrimSpace(string(reportsJSON)) != "null" && len(reportsJSON) > 0 {
			var report []Report
			err = json.Unmarshal(reportsJSON, &report)
			if err != nil {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  err.Error(),
				})
				return
			}

			if report[0].Banned {
				ctx.Notifications().New(app.Notification{
					Title: "Notice",
					Body:  "You have been banned from the platform for theft or abuse.",
				})

				ctx.Navigate("/not-found")
				return
			} else if report[0].Warned {
				ctx.Notifications().New(app.Notification{
					Title: "Notice",
					Body:  "You have a warning for not completing a delivery. Next time you will be banned from the platform.",
				})
			}
		}

		ctx.Dispatch(func(ctx app.Context) {
			h.getDeliveries(ctx)
		})
	})

}

func (h *home) getDeliveries(ctx app.Context) {
	ctx.Async(func() {
		deliveryJSON, err := h.sh.OrbitDocsQuery(dbDelivery, "all", "")
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		if strings.TrimSpace(string(deliveryJSON)) != "null" && len(deliveryJSON) > 0 {
			err = json.Unmarshal(deliveryJSON, &h.deliveries)
			if err != nil {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  err.Error(),
				})
				return
			}

			currentTime := time.Now()

			// Get the date part as a string
			dateTimeString := currentTime.Format("2006-01-02T15:04")

			newSlice := h.deliveries[:0] // reuse underlying array
			for _, d := range h.deliveries {
				if d.Status == StatusDelivered {
					if h.myPeerID == d.CourierID {
						// get reward
						ctx.Notifications().New(app.Notification{
							Title: "Reward",
							Body:  "You have completed delivery with ID: " + d.ID + ". Choose your special item.",
						})

						ctx.Navigate("/delivery/" + d.ID)
						return
					}
				}
				if d.Status == StatusNotDelivered {
					ctx.Async(func() {
						// delete delivered and not delivered ones
						err = h.sh.OrbitDocsDelete(dbDelivery, d.ID)
						if err != nil {
							ctx.Notifications().New(app.Notification{
								Title: "Error",
								Body:  err.Error(),
							})
							return
						}
					})
				} else if d.Status == StatusReady {
					if h.myPeerID == d.CreatedBy {
						ctx.Notifications().New(app.Notification{
							Title: "Delivery in progress",
							Body:  "Courier with ID: " + d.CourierID + " is going to meet you. Check map for route details.",
						})
					} else if slices.Contains(d.Participants, h.myPeerID) {
						ctx.Notifications().New(app.Notification{
							Title: "Delivery in progress",
							Body:  "You won a  competition, you have to complete delivery at: " + d.DeliveryDateTime + ". Check map for route details.",
						})
					}

					newSlice = append(newSlice, d)
				} else if dateTimeString > d.EntryDateTime && d.Status == StatusPending {
					ctx.Async(func() {
						err = h.sh.OrbitDocsDelete(dbDelivery, d.ID)
						if err != nil {
							ctx.Notifications().New(app.Notification{
								Title: "Error",
								Body:  err.Error(),
							})
							return
						}
					})
				} else if dateTimeString <= d.EntryDateTime && slices.Contains(d.Participants, h.myPeerID) && d.Status == StatusPending {
					// enter competition

					// Define the layout for parsing datetime
					layout := "2006-01-02T15:04"

					// Parse the times
					t1, err1 := time.Parse(layout, dateTimeString)
					t2, err2 := time.Parse(layout, d.EntryDateTime)

					if err1 != nil || err2 != nil {
						ctx.Notifications().New(app.Notification{
							Title: "Error",
							Body:  "Error parsing time strings",
						})
						return
					}

					// Calculate the difference (t1 - t2)
					diff := t2.Sub(t1)

					// Convert duration to minutes (or hours, seconds, etc.)
					seconds := diff.Seconds()

					ctx.Async(func() {
						// Create a timer that fires after 5 seconds
						timer := time.NewTimer(time.Duration(seconds) * time.Second)

						// Wait for the timer's channel to send the time (blocks until timer fires)
						<-timer.C

						ctx.Dispatch(func(ctx app.Context) {
							ctx.Navigate("/comp/" + d.ID)
						})
					})

					ctx.Notifications().New(app.Notification{
						Title: "Notice",
						Body:  "You have an upcoming competition at: " + d.EntryDateTime + ". You will be redirected automatically.",
					})

					newSlice = append(newSlice, d)
				} else {
					newSlice = append(newSlice, d)
				}
			}

			h.deliveries = newSlice

			if len(h.deliveries) > 0 {
				deliveryJSON, err = json.Marshal(h.deliveries)
				if err != nil {
					ctx.Notifications().New(app.Notification{
						Title: "Error",
						Body:  err.Error(),
					})
					return
				}

				ctx.Dispatch(func(ctx app.Context) {
					h.deliveryJSON = string(deliveryJSON)
				})
			}
		}
	})
}

// The Render method is where the component appearance is defined.
func (h *home) Render() app.UI {
	return app.Div().ID("home").Body(
		app.Div().ID("header").Body(
			app.H1().Text("Cyber Dérive"),
			newRules(),
		),
		app.Div().Class("links-container").Body(
			app.A().Href("/map").Class("gigantic-link").Text("Request Delivery").OnClick(h.setActionRequest),
			app.Div().Class("vl"),
			app.A().Href("/map").Class("gigantic-link").Text("Make Delivery").OnClick(h.setActionDeliver),
		),
	)
}

func (h *home) setActionRequest(ctx app.Context, e app.Event) {
	ctx.SetState("action", "request")

	var ownDeliveriesSlice []Delivery
	if len(h.deliveries) > 0 {
		// get own delivery requests only
		for _, d := range h.deliveries {
			if d.CreatedBy == h.myPeerID {
				ownDeliveriesSlice = append(ownDeliveriesSlice, d)
			}
		}

		deliveryJSON, err := json.Marshal(ownDeliveriesSlice)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		ctx.SetState("deliveries", string(deliveryJSON))
	}
}

func (h *home) setActionDeliver(ctx app.Context, e app.Event) {
	ctx.SetState("action", "deliver")

	if len(h.deliveries) > 0 {
		var inComp bool
		var inCompSlice []Delivery
		var pendingSlice []Delivery

		// if there is a competition with status in progress and you are not a participant delete from deliveries
		for _, d := range h.deliveries {
			switch d.Status {
			case StatusInCompetition, StatusReady:
				for _, p := range d.Participants {
					if h.myPeerID == p {
						// if in comp keep only that delivery in the array
						inCompSlice = append(inCompSlice, d)
						inComp = true
					}
				}
			case StatusPending:
				// get only non-owned deliveries
				if d.CreatedBy != h.myPeerID {
					pendingSlice = append(pendingSlice, d)
				}
			}
		}

		if inComp {
			deliveryJSON, err := json.Marshal(inCompSlice)
			if err != nil {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  err.Error(),
				})
				return
			}

			ctx.SetState("deliveries", string(deliveryJSON))
		} else {
			deliveryJSON, err := json.Marshal(pendingSlice)
			if err != nil {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  err.Error(),
				})
				return
			}

			ctx.SetState("deliveries", string(deliveryJSON))
		}
	}
}
