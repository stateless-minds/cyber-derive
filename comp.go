package main

import (
	"encoding/json"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	shell "github.com/stateless-minds/go-ipfs-api"
)

// comp is the main component that displays the comp menu. A component is a
// customizable, independent, and reusable UI elemenh. It is created by
// embedding app.Compo into a struch.
type comp struct {
	app.Compo
	sh         *shell.Shell
	sub        *shell.PubSubSubscription
	deliveryID string
	delivery   Delivery
	myPeerID   string
	avatars    []Avatar
	page       string
}

const topicJoinCompetition = "join-competition"

func (c *comp) OnMount(ctx app.Context) {
	sh := shell.NewShell("localhost:5001")
	c.sh = sh

	myPeer, err := c.sh.ID()
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
	}

	ctx.GetState("page", &c.page)

	if len(c.page) == 0 || c.page != "active" {
		ctx.Navigate("/not-found")
		return
	}

	ctx.DelState("page")

	// remove c1 after testing
	c.myPeerID = myPeer.ID + "c1"

	c.subscribeToJoinCompetitionTopic(ctx)

	// remove after testing
	// c.getAvatars(ctx)

	c.setupDeliveryUpdatedListener(ctx)

	path := ctx.Page().URL().Path

	id := strings.TrimPrefix(path, "/comp/")

	c.deliveryID = id

	delivery, err := c.getDelivery()
	if !reflect.DeepEqual(delivery, Delivery{}) && err == nil {
		c.delivery = delivery
		c.deliveryID = id

		var inList bool

		for _, p := range delivery.Participants {
			if p == c.myPeerID {
				inList = true
				delivery.Participants = []string{}
				delivery.Participants = append(delivery.Participants, c.myPeerID)
				delivery.Status = StatusInCompetition
			}
		}

		if !inList {
			delivery.Participants = append(delivery.Participants, c.myPeerID)
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.saveCompetition(ctx, delivery, true)
		})
	} else {
		ctx.Navigate("/not-found")
		return
	}

	app.Window().Call("setupCompetition", true)

}

func (c *comp) setupDeliveryUpdatedListener(ctx app.Context) {
	app.Window().GetElementByID("comp").Call("addEventListener", "delivery-updated", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		event := args[0]

		courierID, err := strconv.Unquote(event.Get("detail").Get("courierId").String())
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return nil
		}

		if c.myPeerID == courierID {
			c.delivery.CourierID = courierID
			c.delivery.Status = StatusReady

			c.saveCompetition(ctx, c.delivery, false)

			ctx.Notifications().New(app.Notification{
				Title: "Competition won",
				Body:  "Congratulations, you won the  competition, you have to complete delivery at: " + c.delivery.DeliveryDateTime + ".",
			})
		}

		return nil
	}))
}

func (c *comp) getDelivery() (Delivery, error) {
	// check if the competition exists
	deliveryJSON, err := c.sh.OrbitDocsGet(dbDelivery, c.deliveryID)
	if err != nil {
		return Delivery{}, err
	}

	var delivery []Delivery

	// if it exists add user to participants
	if strings.TrimSpace(string(deliveryJSON)) != "null" && len(deliveryJSON) > 0 {
		err = json.Unmarshal(deliveryJSON, &delivery)
		if err != nil {
			return Delivery{}, err
		}
	} else {
		return Delivery{}, err
	}

	return delivery[0], nil
}

func (c *comp) saveCompetition(ctx app.Context, delivery Delivery, publish bool) {
	deliveryJSON, err := json.Marshal(delivery)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	ctx.Async(func() {
		err = c.sh.OrbitDocsPut(dbDelivery, deliveryJSON)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		if publish {
			err = c.sh.PubSubPublish(topicJoinCompetition, c.myPeerID)
			if err != nil {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  err.Error(),
				})
				return
			}
		}
	})
}

func (c *comp) getAvatars(ctx app.Context) {
	// var avatars []Avatar
	ctx.Async(func() {
		avatarJSON, err := c.sh.OrbitDocsQuery(dbAvatar, "all", "")
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		if strings.TrimSpace(string(avatarJSON)) != "null" && len(avatarJSON) > 0 {

			var avatars []Avatar

			err = json.Unmarshal(avatarJSON, &avatars)
			if err != nil {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  err.Error(),
				})
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				for _, avatar := range avatars {
					c.avatars = append(c.avatars, avatar)
					err = c.sh.PubSubPublish(topicJoinCompetition, avatar.ID)
					if err != nil {
						ctx.Notifications().New(app.Notification{
							Title: "Error",
							Body:  err.Error(),
						})
						return
					}
				}
			})
		} else {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  "Avatar not found",
			})
			return
		}
	})
}

func (c *comp) getAvatar(ctx app.Context, peerID string) {
	// var avatars []Avatar
	ctx.Async(func() {
		avatarJSON, err := c.sh.OrbitDocsGet(dbAvatar, peerID)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		if strings.TrimSpace(string(avatarJSON)) != "null" && len(avatarJSON) > 0 {

			var avatars []Avatar

			err = json.Unmarshal(avatarJSON, &avatars)
			if err != nil {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  err.Error(),
				})
				return
			}

			// avatars = append(avatars, avatar)

			ctx.Dispatch(func(ctx app.Context) {
				c.avatars = append(c.avatars, avatars[0])
			})
		} else {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  "Avatar not found",
			})
			return
		}
	})
}

func (c *comp) subscribeToJoinCompetitionTopic(ctx app.Context) {
	ctx.Async(func() {
		topic := topicJoinCompetition
		subscription, err := c.sh.PubSubSubscribe(topic)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  "Could not subscribe to topic",
			})
			return
		}
		c.sub = subscription
		c.listenForJoins(ctx)
	})
}

func (c *comp) listenForJoins(ctx app.Context) {
	timeoutDuration := 10 * time.Second
	timeoutTimer := time.NewTimer(timeoutDuration)

	msgChan := make(chan *shell.Message)
	errChan := make(chan error)

	ctx.Async(func() {
		for {
			ctx.Async(func() {
				msg, err := c.sub.Next()
				if err != nil {
					errChan <- err
					return
				}

				ctx.Dispatch(func(ctx app.Context) {
					msgChan <- msg
				})
			})

			select {
			case <-timeoutTimer.C:
				c.sub.Cancel()
				app.Window().Call("setupCompetition", false)
				return

			case err := <-errChan:
				if err == io.EOF {
					app.Log("Subscription closed, exiting listener")
				} else {
					app.Log("Error reading from subscription: %v", err)
				}
				return // exit the loop and goroutine on EOF or error

			case msg := <-msgChan:
				// remove after testing
				c.getAvatar(ctx, msg.From.String()+"c1")
				// c.getAvatar(ctx, msg.From.String())

				// Reset the timer because a message arrived
				if !timeoutTimer.Stop() {
					<-timeoutTimer.C
				}
				timeoutTimer.Reset(timeoutDuration)
			}
		}
	})
}

// The Render method is where the component appearance is defined.
func (c *comp) Render() app.UI {
	return app.Div().ID("container").Body(
		app.Div().Class("overlay").Body(
			app.Span().ID("countdown").Class("countdown-text"),
		),
		app.Div().ID("header").Body(
			app.P().Body(
				app.H1().Text("Cyber Dérive"),
				newRules(),
				app.Span().Class("comp-title").Text("Competition ID: "+c.deliveryID),
				app.Span().Class("own-id").Text("Your Player ID: "+c.myPeerID),
			),
		),
		app.Div().ID("comp").Body(
			app.Range(c.avatars).Slice(func(i int) app.UI {
				return app.If(c.myPeerID == c.avatars[i].ID, func() app.UI {
					return app.Div().ID(c.avatars[i].ID).Class("comp-item own-item").Body(
						app.P().Class("peer-id").Body(
							app.If(c.myPeerID == c.avatars[i].ID, func() app.UI {
								return app.Span().Class("label-id").Text("You: ")
							}),
						),
						app.Img().Class("comp-avatar-img").Src("web/assets/body-4444.jpeg"),
						app.Div().Class("stats-container").Body(
							app.Label().Class("label-stats").Text(" Energy: "),
							app.Span().ID("energy").Class("span-stats").Text(c.avatars[i].Energy),
							app.Label().Class("label-stats").Text(" Speed: "),
							app.Span().ID("speed").Class("span-stats").Text(c.avatars[i].Speed),
							app.Label().Class("label-stats").Text(" Safety: "),
							app.Span().ID("safety").Class("span-stats").Text(c.avatars[i].Safety),
						),
					)
				}).Else(func() app.UI {
					return app.Div().ID(c.avatars[i].ID).Class("comp-item").Body(
						app.P().Class("peer-id").Body(
							app.P().Body(
								app.Span().Class("label-id").Text("Player ID: "),
								app.Span().Class("player-id").Text(c.avatars[i].ID),
							),
						),
						app.Img().Class("comp-avatar-img").Src("web/assets/body-4444.jpeg"),
						app.Div().Class("stats-container").Body(
							app.Label().Class("label-stats").Text(" Energy: "),
							app.Span().ID("energy").Class("span-stats").Text(c.avatars[i].Energy),
							app.Label().Class("label-stats").Text(" Speed: "),
							app.Span().ID("speed").Class("span-stats").Text(c.avatars[i].Speed),
							app.Label().Class("label-stats").Text(" Safety: "),
							app.Span().ID("safety").Class("span-stats").Text(c.avatars[i].Safety),
						),
					)
				})
			}),
		),
	)
}
