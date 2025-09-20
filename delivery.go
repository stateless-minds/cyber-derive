package main

import (
	"encoding/json"
	"hash/fnv"
	"math/rand"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	shell "github.com/stateless-minds/go-ipfs-api"
)

const dbCard = "card"
const dbAvatar = "avatar"

// card is the main component that displays the initial 5 cards. A component is a
// customizable, independent, and reusable UI elemenh. It is created by
// embedding app.Compo into a struch.
type delivery struct {
	app.Compo
	sh                  *shell.Shell
	delivery            Delivery
	deliveryID          string
	myPeerID            string
	avatar              Avatar
	allCards            []Card
	cardDeck            []Card
	equippedCards       []Card
	processedHandItem   bool
	processedHandItemID string
	special             string
	page                string
}

type Card struct {
	ID       string `mapstructure:"_id" json:"_id" validate:"uuid_rfc4122"`           // ID
	Category string `mapstructure:"category" json:"category" validate:"uuid_rfc4122"` // Category
	Image    string `mapstructure:"image" json:"image" validate:"uuid_rfc4122"`       // Image base64
	Stat     string `mapstructure:"stat" json:"stat" validate:"uuid_rfc4122"`         // Stat
	Value    int    `mapstructure:"value" json:"value" validate:"uuid_rfc4122"`       // Value
	Special  string `mapstructure:"special" json:"special" validate:"uuid_rfc4122"`   // Special
}

type Avatar struct {
	ID     string   `mapstructure:"_id" json:"_id" validate:"uuid_rfc4122"`       // ID
	Energy int      `mapstructure:"energy" json:"energy" validate:"uuid_rfc4122"` // Energy
	Safety int      `mapstructure:"safety" json:"safety" validate:"uuid_rfc4122"` // Safety
	Speed  int      `mapstructure:"speed" json:"speed" validate:"uuid_rfc4122"`   // Speed
	Cards  []string `mapstructure:"cards" json:"cards" validate:"uuid_rfc4122"`   // Cards
}

func (d *delivery) OnMount(ctx app.Context) {
	sh := shell.NewShell("localhost:5001")
	d.sh = sh

	myPeer, err := d.sh.ID()
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
	}

	ctx.GetState("page", &d.page)

	if len(d.page) == 0 || d.page != "active" {
		ctx.Navigate("/not-found")
		return
	}

	ctx.DelState("page")

	// remove c1 after testing
	d.myPeerID = myPeer.ID + "c1"

	path := ctx.Page().URL().Path

	id := strings.TrimPrefix(path, "/delivery/")

	d.deliveryID = id

	// enable after testing

	delivery, err := d.deliveryExists()

	if !reflect.DeepEqual(delivery, Delivery{}) && err == nil {
		d.delivery = delivery
	} else {
		ctx.Navigate("/not-found")
		return
	}

	if d.delivery.Status == StatusDelivered {
		d.special = "yes"
	}

	d.getCards(ctx)
}

func (d *delivery) getCards(ctx app.Context) {
	// get 16 random cards
	ctx.Async(func() {
		cardsJSON, err := d.sh.OrbitDocsQuery(dbCard, "all", "")
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
		}

		var cards []Card
		var randomCards []Card

		if len(cardsJSON) > 0 {
			err = json.Unmarshal(cardsJSON, &cards)
			if err != nil {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  err.Error(),
				})
			}

			d.allCards = cards

			seed := hashStringToSeed(d.deliveryID)
			r := rand.New(rand.NewSource(seed)) // deterministic RNG seeded by delivery ID

			var min, max int
			if d.special == "yes" {
				min = 1001
				max = 1085
			} else {
				min = 1
				max = 96
			}

			// Generate a slice with all numbers in range [min, max]
			nums := make([]int, max-min+1)
			for i := range nums {
				nums[i] = min + i
			}

			// Shuffle the slice using deterministic rng
			r.Shuffle(len(nums), func(i, j int) {
				nums[i], nums[j] = nums[j], nums[i]
			})

			// Take the first 12 unique numbers
			result := nums[:12]

			for _, rn := range result {
				for _, card := range cards {
					if card.ID == strconv.Itoa(rn) {
						randomCards = append(randomCards, card)
					}
				}
			}

			ctx.Dispatch(func(ctx app.Context) {
				d.cardDeck = randomCards
				d.getAvatar(ctx)
			})
		}
	})
}

// hashStringToSeed hashes a string delivery ID into a uint64 seed
func hashStringToSeed(s string) int64 {
	h := fnv.New64()
	h.Write([]byte(s))
	return int64(h.Sum64())
}

func (d *delivery) getAvatar(ctx app.Context) {
	// get avatar
	ctx.Async(func() {
		avatarJSON, err := d.sh.OrbitDocsGet(dbAvatar, d.myPeerID)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
		}

		var avatar []Avatar

		if strings.TrimSpace(string(avatarJSON)) != "null" && len(avatarJSON) > 0 {
			err = json.Unmarshal(avatarJSON, &avatar)
			if err != nil {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  err.Error(),
				})
			}

			var equippedCards []Card

			for _, card := range d.allCards {
				if slices.Contains(avatar[0].Cards, card.ID) {
					equippedCards = append(equippedCards, card)
					d.cardDeck = removeItem(d.cardDeck, card)
				}
			}

			ctx.Dispatch(func(ctx app.Context) {
				d.avatar = avatar[0]
				d.equippedCards = equippedCards
			})
		}
		app.Window().Call("setupAvatar")
	})
}

func removeItem(slice []Card, item Card) []Card {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice // item not found, return original slice
}

func (d *delivery) deleteDelivery(ctx app.Context) {
	ctx.Async(func() {
		err := d.sh.OrbitDocsDelete(dbDelivery, d.deliveryID)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
		}

		ctx.Dispatch(func(ctx app.Context) {
			ctx.Navigate("/")
		})
	})
}

// The Render method is where the component appearance is defined.
func (d *delivery) Render() app.UI {
	return app.Div().ID("container").Body(
		app.Div().ID("header").Body(
			app.H1().Text("Cyber Dérive"),
			newRules(),
		),
		app.Div().ID("inventory").Body(
			app.Range(d.cardDeck).Slice(func(i int) app.UI {
				return app.Div().ID("card-"+d.cardDeck[i].ID).Class("card").Draggable(true).DataSet("category", d.cardDeck[i].Category).DataSet("stat", d.cardDeck[i].Stat).DataSet("value", d.cardDeck[i].Value).TabIndex(0).Aria("grabbed", false).Aria("roledescription", "draggable item").Body(
					app.Img().Src("data:image/jpeg;base64,"+d.cardDeck[i].Image),
					app.Div().Class("text-overlay").Text(d.cardDeck[i].Stat+" "+strconv.Itoa(d.cardDeck[i].Value)),
					app.Span().Class("drop-label").Text("Drag and drop"),
				)
			}),
			app.Div().Class("vl compact"),
		),
		app.Div().ID("avatar").Aria("label", "Avatar equipment slots").Body(
			app.Label().Class("label-stats").Text("Energy: "),
			app.Span().ID("energy").Class("span-stats").Text(d.avatar.Energy),
			app.Label().Class("label-stats").Text("Safety: "),
			app.Span().ID("safety").Class("span-stats").Text(d.avatar.Safety),
			app.Label().Class("label-stats").Text("Speed: "),
			app.Span().ID("speed").Class("span-stats").Text(d.avatar.Speed),
			app.Div().ID("slot-head").Class("slot").DataSet("slot", "head").DataSet("accept", "head").Aria("dropeffect", "move").TabIndex(0).Aria("label", "Head slot").Body(
				app.Range(d.equippedCards).Slice(func(i int) app.UI {
					return app.If(d.equippedCards[i].Category == "head", func() app.UI {
						return app.Div().ID("card-"+d.equippedCards[i].ID).Class("card").Draggable(true).DataSet("category", d.equippedCards[i].Category).DataSet("stat", d.equippedCards[i].Stat).DataSet("value", d.equippedCards[i].Value).DataSet("equipped", true).TabIndex(0).Aria("grabbed", false).Aria("roledescription", "draggable item").Body(
							app.Img().Src("data:image/jpeg;base64,"+d.equippedCards[i].Image),
							app.Div().Class("text-overlay").Text(d.equippedCards[i].Stat+" "+strconv.Itoa(d.equippedCards[i].Value)),
							app.Span().Class("drop-label").Text("Drag and drop"),
						)
					})
				}),
			),
			app.Div().ID("slot-left-hand").Class("slot").DataSet("slot", "left-hand").DataSet("accept", "hands").Aria("dropeffect", "move").TabIndex(0).Aria("label", "Left hand slot").Body(
				app.Range(d.equippedCards).Slice(func(i int) app.UI {
					return app.If(d.equippedCards[i].Category == "hands" && !d.processedHandItem, func() app.UI {
						d.processedHandItemID = d.equippedCards[i].ID
						d.processedHandItem = true
						return app.Div().ID("card-"+d.equippedCards[i].ID).Class("card").Draggable(true).DataSet("category", d.equippedCards[i].Category).DataSet("stat", d.equippedCards[i].Stat).DataSet("value", d.equippedCards[i].Value).DataSet("equipped", true).TabIndex(0).Aria("grabbed", false).Aria("roledescription", "draggable item").Body(
							app.Img().Src("data:image/jpeg;base64,"+d.equippedCards[i].Image),
							app.Div().Class("text-overlay").Text(d.equippedCards[i].Stat+" "+strconv.Itoa(d.equippedCards[i].Value)),
							app.Span().Class("drop-label").Text("Drag and drop"),
						)
					})
				}),
			),
			app.Div().ID("slot-core").Class("slot").DataSet("slot", "core").DataSet("accept", "core").Aria("dropeffect", "move").TabIndex(0).Aria("label", "core slot").Body(
				app.Range(d.equippedCards).Slice(func(i int) app.UI {
					return app.If(d.equippedCards[i].Category == "core", func() app.UI {
						return app.Div().ID("card-"+d.equippedCards[i].ID).Class("card").Draggable(true).DataSet("category", d.equippedCards[i].Category).DataSet("stat", d.equippedCards[i].Stat).DataSet("value", d.equippedCards[i].Value).DataSet("equipped", true).TabIndex(0).Aria("grabbed", false).Aria("roledescription", "draggable item").Body(
							app.Img().Src("data:image/jpeg;base64,"+d.equippedCards[i].Image),
							app.Div().Class("text-overlay").Text(d.equippedCards[i].Stat+" "+strconv.Itoa(d.equippedCards[i].Value)),
							app.Span().Class("drop-label").Text("Drag and drop"),
						)
					})
				}),
			),
			app.Div().ID("slot-right-hand").Class("slot").DataSet("slot", "right-hand").DataSet("accept", "hands").Aria("dropeffect", "move").TabIndex(0).Aria("label", "Right hand slot").Body(
				app.Range(d.equippedCards).Slice(func(i int) app.UI {
					return app.If(d.equippedCards[i].Category == "hands" && d.equippedCards[i].ID != d.processedHandItemID, func() app.UI {
						d.processedHandItem = false
						return app.Div().ID("card-"+d.equippedCards[i].ID).Class("card").Draggable(true).DataSet("category", d.equippedCards[i].Category).DataSet("stat", d.equippedCards[i].Stat).DataSet("value", d.equippedCards[i].Value).DataSet("equipped", true).TabIndex(0).Aria("grabbed", false).Aria("roledescription", "draggable item").Body(
							app.Img().Src("data:image/jpeg;base64,"+d.equippedCards[i].Image),
							app.Div().Class("text-overlay").Text(d.equippedCards[i].Stat+" "+strconv.Itoa(d.equippedCards[i].Value)),
							app.Span().Class("drop-label").Text("Drag and drop"),
						)
					})
				}),
			),
			app.Div().ID("slot-legs").Class("slot").DataSet("slot", "legs").DataSet("accept", "legs").Aria("dropeffect", "move").TabIndex(0).Aria("label", "Legs slot").Body(
				app.Range(d.equippedCards).Slice(func(i int) app.UI {
					return app.If(d.equippedCards[i].Category == "legs", func() app.UI {
						return app.Div().ID("card-"+d.equippedCards[i].ID).Class("card").Draggable(true).DataSet("category", d.equippedCards[i].Category).DataSet("stat", d.equippedCards[i].Stat).DataSet("value", d.equippedCards[i].Value).DataSet("equipped", true).TabIndex(0).Aria("grabbed", false).Aria("roledescription", "draggable item").Body(
							app.Img().Src("data:image/jpeg;base64,"+d.equippedCards[i].Image),
							app.Div().Class("text-overlay").Text(d.equippedCards[i].Stat+" "+strconv.Itoa(d.equippedCards[i].Value)),
							app.Span().Class("drop-label").Text("Drag and drop"),
						)
					})
				}),
			),
			app.If(d.special == "yes", func() app.UI {
				return app.Button().ID("save-avatar").Text("Submit").Disabled(true).Value(len(d.equippedCards)).OnClick(d.saveAvatar)
			}).Else(func() app.UI {
				return app.Button().ID("save-avatar").Text("Enter Competition").Disabled(true).Value(len(d.equippedCards)).OnClick(d.saveAvatar)
			}),
		),
		app.Div().ID("tooltip").Styles(map[string]string{
			"display":        "none",
			"position":       "fixed",
			"z-index":        "1000",
			"padding":        "8px 12px",
			"background":     "#333",
			"color":          "#fff",
			"border-radius":  "5px",
			"font-size":      "14px",
			"box-shadow":     "0 2px 8px #222",
			"pointer-events": "none",
		}),
	)
}

func (d *delivery) saveAvatar(ctx app.Context, e app.Event) {
	energy := app.Window().GetElementByID("energy").Get("textContent").String()
	energyInt, err := strconv.Atoi(energy)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
	}

	safety := app.Window().GetElementByID("safety").Get("textContent").String()
	safetyInt, err := strconv.Atoi(safety)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
	}

	speed := app.Window().GetElementByID("speed").Get("textContent").String()
	speedInt, err := strconv.Atoi(speed)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
	}

	var cards []string

	head := app.Window().GetElementByID("slot-head")

	// Get the children HTMLCollection
	children := head.Get("children")

	// Get the first child (index 0)
	if children.Get("length").Int() > 0 {
		firstChild := children.Index(0)
		// Example: get its text content
		id := firstChild.Get("id").String()
		headCardID := strings.TrimPrefix(id, "card-") // Strip "card-"
		cards = append(cards, headCardID)
	}

	leftHand := app.Window().GetElementByID("slot-left-hand")

	// Get the children HTMLCollection
	children = leftHand.Get("children")

	// Get the first child (index 0)
	if children.Get("length").Int() > 0 {
		firstChild := children.Index(0)
		// Example: get its text content
		id := firstChild.Get("id").String()
		leftHandCardID := strings.TrimPrefix(id, "card-") // Strip "card-"
		cards = append(cards, leftHandCardID)
	}

	rightHand := app.Window().GetElementByID("slot-right-hand")

	// Get the children HTMLCollection
	children = rightHand.Get("children")

	// Get the first child (index 0)
	if children.Get("length").Int() > 0 {
		firstChild := children.Index(0)
		// Example: get its text content
		id := firstChild.Get("id").String()
		rightHandCardID := strings.TrimPrefix(id, "card-") // Strip "card-"
		cards = append(cards, rightHandCardID)
	}

	core := app.Window().GetElementByID("slot-core")

	// Get the children HTMLCollection
	children = core.Get("children")

	// Get the first child (index 0)
	if children.Get("length").Int() > 0 {
		firstChild := children.Index(0)
		// Example: get its text content
		id := firstChild.Get("id").String()
		coreCardID := strings.TrimPrefix(id, "card-") // Strip "card-"
		cards = append(cards, coreCardID)
	}

	legs := app.Window().GetElementByID("slot-legs")
	// Get the children HTMLCollection
	children = legs.Get("children")

	// Get the first child (index 0)
	if children.Get("length").Int() > 0 {
		firstChild := children.Index(0)
		// Example: get its text content
		id := firstChild.Get("id").String()
		legsCardID := strings.TrimPrefix(id, "card-") // Strip "card-"
		cards = append(cards, legsCardID)
	}

	avatar := Avatar{
		ID:     d.myPeerID,
		Energy: energyInt,
		Safety: safetyInt,
		Speed:  speedInt,
		Cards:  cards,
	}

	avatarJSON, err := json.Marshal(avatar)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	ctx.Async(func() {
		err = d.sh.OrbitDocsPut(dbAvatar, avatarJSON)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			if d.special == "yes" {
				d.deleteDelivery(ctx)
			} else {
				d.saveDelivery(ctx)
			}
		})
	})
}

func (d *delivery) deliveryExists() (Delivery, error) {
	// check if the competition exists
	deliveryJSON, err := d.sh.OrbitDocsGet(dbDelivery, d.deliveryID)
	if err != nil {
		return Delivery{}, err
	}

	var delivery []Delivery

	if strings.TrimSpace(string(deliveryJSON)) != "null" && len(deliveryJSON) > 0 {
		err = json.Unmarshal(deliveryJSON, &delivery)
		if err != nil {
			return Delivery{}, err
		}
	} else {
		return Delivery{}, nil
	}

	return delivery[0], nil
}

func (d *delivery) saveDelivery(ctx app.Context) {
	d.delivery.Participants = append(d.delivery.Participants, d.myPeerID)

	deliveryJSON, err := json.Marshal(d.delivery)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	ctx.Async(func() {
		err = d.sh.OrbitDocsPut(dbDelivery, deliveryJSON)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			ctx.Navigate("/")
		})
	})
}
