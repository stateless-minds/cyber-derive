package main

import (
	"encoding/json"
	"strconv"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	shell "github.com/stateless-minds/go-ipfs-api"
)

const dbDelivery = "delivery"

type Status string

const (
	StatusPending       Status = "pending"
	StatusInCompetition Status = "in_competition"
	StatusReady         Status = "ready_to_deliver"
	StatusDelivered     Status = "delivered"
	StatusNotDelivered  Status = "not_delivered"
)

// delivery is the delivery component that displays a delivery on a map. A component is a
// customizable, independent, and reusable UI elemend. It is created by
// embedding app.Compo into a strucd.
type mapLibre struct {
	app.Compo
	sh             *shell.Shell
	action         string
	deliveriesJSON string
	myPeerID       string
}

type Delivery struct {
	ID               string     `mapstructure:"_id" json:"_id" validate:"uuid_rfc4122"`                             // Unique identifier for the delivery
	Coordinates      [][]string `mapstructure:"coordinates" json:"coordinates" validate:"uuid_rfc4122"`             // Coordinates
	DeliveryDateTime string     `mapstructure:"delivery_datetime" json:"delivery_datetime" validate:"uuid_rfc4122"` // Delivery DateTime
	EntryDateTime    string     `mapstructure:"entry_datetime" json:"entry_datetime" validate:"uuid_rfc4122"`       // Entry DateTime
	CreatedBy        string     `mapstructure:"created_by" json:"created_by" validate:"uuid_rfc4122"`               // Created By
	Participants     []string   `mapstructure:"participants" json:"participants" validate:"uuid_rfc4122"`           // Participants
	CourierID        string     `mapstructure:"courier_id" json:"courier_id" validate:"uuid_rfc4122"`               // CourierID of competition winner
	Status           Status     `mapstructure:"status" json:"status" validate:"uuid_rfc4122"`                       // Status
}

func (m *mapLibre) OnMount(ctx app.Context) {
	sh := shell.NewShell("localhost:5001")
	m.sh = sh

	myPeer, err := m.sh.ID()
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
	}

	// remove c1 after testing
	m.myPeerID = myPeer.ID + "c1"

	ctx.GetState("action", &m.action)

	if m.action == "" {
		ctx.Navigate("/not-found")
		return
	}

	ctx.GetState("deliveries", &m.deliveriesJSON)

	app.Window().Call("setupMap", m.myPeerID, m.action, m.deliveriesJSON)

	m.setupDeliveryCreatedListener(ctx)

	m.setupDeliveryCompletedListener(ctx)

	m.setupDeliveryFailedListener(ctx)

	m.setupDeliveryRedirectListener(ctx)
}

func (m *mapLibre) setupDeliveryCreatedListener(ctx app.Context) {
	app.Window().GetElementByID("map").Call("addEventListener", "delivery-created", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		event := args[0]

		deliveryID, err := strconv.Unquote(event.Get("detail").Get("id").String())
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return nil
		}

		dc := event.Get("detail").Get("latlngs").String()

		// Now unmarshal the JSON from the cleaned string
		var deliveryCoordinates [][]string
		err = json.Unmarshal([]byte(dc), &deliveryCoordinates)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return nil
		}

		deliveryDateTime, err := strconv.Unquote(event.Get("detail").Get("deliveryDateTime").String())
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return nil
		}

		entryDateTime, err := strconv.Unquote(event.Get("detail").Get("entryDateTime").String())
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return nil
		}

		delivery := Delivery{
			ID:               deliveryID,
			Coordinates:      deliveryCoordinates,
			DeliveryDateTime: deliveryDateTime,
			EntryDateTime:    entryDateTime,
			CreatedBy:        m.myPeerID,
			Participants:     []string{},
			Status:           StatusPending,
		}

		deliveryJSON, err := json.Marshal(delivery)

		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
		}

		ctx.Async(func() {
			err = m.sh.OrbitDocsPut(dbDelivery, deliveryJSON)
			if err != nil {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  err.Error(),
				})
			}

			ctx.Dispatch(func(ctx app.Context) {
				ctx.Notifications().New(app.Notification{
					Title: "Success",
					Body:  "Delivery was saved. Come back at Enter by time to check who is going to deliver.",
				})

				ctx.Navigate("/")
			})
		})

		return nil
	}))
}

func (m *mapLibre) setupDeliveryCompletedListener(ctx app.Context) {
	app.Window().GetElementByID("map").Call("addEventListener", "delivery-completed", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		event := args[0]

		deliveryID, err := strconv.Unquote(event.Get("detail").Get("id").String())
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return nil
		}

		var deliveries []Delivery

		err = json.Unmarshal([]byte(m.deliveriesJSON), &deliveries)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return nil
		}

		for _, delivery := range deliveries {
			if delivery.ID == deliveryID {
				delivery.Status = StatusDelivered
				deliveryJSON, err := json.Marshal(delivery)

				if err != nil {
					ctx.Notifications().New(app.Notification{
						Title: "Error",
						Body:  err.Error(),
					})
				}

				ctx.Async(func() {
					err = m.sh.OrbitDocsPut(dbDelivery, deliveryJSON)
					if err != nil {
						ctx.Notifications().New(app.Notification{
							Title: "Error",
							Body:  err.Error(),
						})
					}

					ctx.Dispatch(func(ctx app.Context) {
						ctx.Notifications().New(app.Notification{
							Title: "Success",
							Body:  "Delivery was completed.",
						})

						ctx.Navigate("/")
					})
				})
			}
		}

		return nil
	}))
}

func (m *mapLibre) setupDeliveryFailedListener(ctx app.Context) {
	app.Window().GetElementByID("map").Call("addEventListener", "delivery-failed", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		event := args[0]

		deliveryID, err := strconv.Unquote(event.Get("detail").Get("id").String())
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return nil
		}

		var deliveries []Delivery

		err = json.Unmarshal([]byte(m.deliveriesJSON), &deliveries)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return nil
		}

		for _, delivery := range deliveries {
			if delivery.ID == deliveryID {
				delivery.Status = StatusNotDelivered
				deliveryJSON, err := json.Marshal(delivery)

				if err != nil {
					ctx.Notifications().New(app.Notification{
						Title: "Error",
						Body:  err.Error(),
					})
				}

				ctx.Async(func() {
					err = m.sh.OrbitDocsPut(dbDelivery, deliveryJSON)
					if err != nil {
						ctx.Notifications().New(app.Notification{
							Title: "Error",
							Body:  err.Error(),
						})
					}

					ctx.Dispatch(func(ctx app.Context) {
						ctx.Notifications().New(app.Notification{
							Title: "Notice",
							Body:  "Please file a report with reason for failed delivery.",
						})

						ctx.SetState("delivery", delivery)

						ctx.Navigate("/report")
					})
				})
			}
		}

		return nil
	}))
}

func (m *mapLibre) setupDeliveryRedirectListener(ctx app.Context) {
	app.Window().GetElementByID("map").Call("addEventListener", "delivery-redirected", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		event := args[0]

		deliveryID, err := strconv.Unquote(event.Get("detail").Get("id").String())
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return nil
		}

		ctx.Navigate("/delivery/" + deliveryID)

		return nil
	}))
}

// The Render method is where the component appearance is defined.
func (m *mapLibre) Render() app.UI {
	return app.Div().ID("map")
}
