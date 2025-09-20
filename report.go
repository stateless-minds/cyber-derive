package main

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	shell "github.com/stateless-minds/go-ipfs-api"
)

const dbReport = "report"

// report is the report component that displays the report interface. A component is a
// customizable, independent, and reusable UI elemenh. It is created by
// embedding app.Compo into a struch.
type report struct {
	app.Compo
	sh          *shell.Shell
	delivery    Delivery
	item        Report
	reason      string
	description string
}

type Reason string

const (
	ReasonTheft      Reason = "theft"
	ReasonAbuse      Reason = "abuse"
	ReasonNotPresent Reason = "not_present"
)

type Report struct {
	ID           string   `mapstructure:"_id" json:"_id" validate:"uuid_rfc4122"`                 // ID
	Reasons      []Reason `mapstructure:"reason" json:"reason" validate:"uuid_rfc4122"`           // Reason
	Descriptions []string `mapstructure:"description" json:"description" validate:"uuid_rfc4122"` // Description
	Reporters    []string `mapstructure:"reporter_id" json:"reporter_id" validate:"uuid_rfc4122"` // ReporterID
	Deliveries   []string `mapstructure:"delivery_id" json:"delivery_id" validate:"uuid_rfc4122"` // ReporterID
	Warned       bool     `mapstructure:"warned" json:"warned" validate:"uuid_rfc4122"`           // Warned
	Banned       bool     `mapstructure:"banned" json:"banned" validate:"uuid_rfc4122"`           // Banned
}

func (r *report) OnMount(ctx app.Context) {
	sh := shell.NewShell("localhost:5001")
	r.sh = sh

	myPeer, err := r.sh.ID()
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
	}

	ctx.GetState("delivery", &r.delivery)

	// if r.delivery.CreatedBy != myPeer.ID {
	// remove after testing
	if r.delivery.CreatedBy != myPeer.ID+"r1" {
		ctx.Navigate("/not-found")
		return
	}

	r.item, err = r.getReport(ctx)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
	}
}

func (r *report) getReport(ctx app.Context) (Report, error) {
	// check if the competition exists
	reportJSON, err := r.sh.OrbitDocsGet(dbReport, r.delivery.CourierID)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return Report{}, errors.New("error getting report")
	}

	var report []Report

	// if it exists add user to participants
	if strings.TrimSpace(string(reportJSON)) != "null" && len(reportJSON) > 0 {
		err = json.Unmarshal(reportJSON, &report)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return Report{}, errors.New("error unmarshalling report")
		}
	} else {
		return Report{}, nil
	}

	return report[0], nil
}

func (r *report) saveReport(ctx app.Context, report Report) {
	reportJSON, err := json.Marshal(report)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	ctx.Async(func() {
		err = r.sh.OrbitDocsPut(dbReport, reportJSON)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			ctx.Notifications().New(app.Notification{
				Title: "Success",
				Body:  "Report has been submitted.",
			})

			ctx.Navigate("/")
		})
	})
}

// The Render method is where the component appearance is defined.
func (r *report) Render() app.UI {
	return app.Div().ID("report").Body(
		app.Div().ID("header").Body(
			app.H1().Text("Cyber Dérive"),
			newRules(),
		),
		app.Div().Class("report-form").Body(
			app.Form().Body(
				app.H2().Text("Create Report"),
				app.Label().For("report-type").Text("Type of Report"),
				app.Select().ID("report-type").Name("report-type").Required(true).Body(
					app.Option().Value("").Disabled(true).Selected(true).Text("Select an option"),
					app.Option().Value("theft").Text("Theft"),
					app.Option().Value("abuse").Text("Abuse"),
					app.Option().Value("not-present").Text("Not Present"),
				).OnChange(r.selectChange),
				app.Label().For("description").Text("Describe the case: "),
				app.Textarea().ID("description").Name("description").Rows(10).Placeholder("Provide details here...").Required(true).OnChange(r.textAreaChange),
				app.Button().Type("submit").Text("Submit").OnClick(r.sendReport),
			),
		),
	)
}

func (r *report) selectChange(ctx app.Context, e app.Event) {
	v := ctx.JSSrc().Get("value").String()
	r.reason = v
}

func (r *report) textAreaChange(ctx app.Context, e app.Event) {
	v := ctx.JSSrc().Get("value").String()
	r.description = v
}

func (r *report) sendReport(ctx app.Context, e app.Event) {
	e.PreventDefault()

	if reflect.DeepEqual(r.item, Report{}) {
		report := Report{
			ID:           r.delivery.CourierID,
			Reasons:      []Reason{Reason(r.reason)},
			Descriptions: []string{r.description},
			Reporters:    []string{r.delivery.CreatedBy},
			Deliveries:   []string{r.delivery.ID},
		}
		if r.reason == string(ReasonTheft) || r.reason == string(ReasonAbuse) {
			report.Banned = true
		} else {
			report.Warned = true
		}

		r.saveReport(ctx, report)
	} else {
		// the user had a warning
		r.item.Banned = true
		r.item.Reasons = append(r.item.Reasons, Reason(r.reason))
		r.item.Reporters = append(r.item.Reporters, r.delivery.CreatedBy)
		r.item.Descriptions = append(r.item.Descriptions, r.description)
		r.item.Deliveries = append(r.item.Deliveries, r.delivery.ID)
		r.saveReport(ctx, r.item)
	}
}
