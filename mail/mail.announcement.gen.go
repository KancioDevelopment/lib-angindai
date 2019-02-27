package mail

import (
	"fmt"
	"time"

	"github.com/matcornic/hermes"
)

type (
	EmailContentAnnouncement struct {
		MimeType     MimeType
		Organization string
		Intros       []string
		Greeting     string
		Signature    string
	}

	Receipt struct {
		MimeType     MimeType
		URL          string
		CustomerName string
		Organization string
		Intros       []string
		Greeting     string
		Signature    string
		TableData    [][]hermes.Entry
		ActionIntro  string
		ButtonText   string
		ButtonLink   string
	}
)

func GenerateAnnouncementEmail(option EmailContentAnnouncement) (string, error) {
	h := hermes.Hermes{
		Product: hermes.Product{
			Name: option.Organization,
			Link: "",
			Copyright: fmt.Sprintf("Copyright © %d %s. All rights reserved.",
				time.Now().Year(), option.Organization),
		},
	}

	email := hermes.Email{
		Body: hermes.Body{
			Title:     option.Greeting,
			Intros:    option.Intros,
			Signature: option.Signature,
		},
	}

	switch option.MimeType {
	case MimeTypeHtml:
		return h.GenerateHTML(email)
	}

	return h.GeneratePlainText(email)
}

func GenerateReceipt(option Receipt) (string, error) {
	h := hermes.Hermes{
		Product: hermes.Product{
			Name: option.Organization,
			Link: option.URL,
			Copyright: fmt.Sprintf("Copyright © %d %s. All rights reserved.",
				time.Now().Year(), option.Organization),
		},
	}

	email := hermes.Email{
		Body: hermes.Body{
			Name:   option.CustomerName,
			Intros: option.Intros,
			Table: hermes.Table{
				Data: option.TableData,
				Columns: hermes.Columns{
					CustomWidth: map[string]string{
						"Item":  "20%",
						"Price": "15%",
					},
					CustomAlignment: map[string]string{
						"Price": "right",
					},
				},
			},
			Actions: []hermes.Action{
				{
					Instructions: option.ActionIntro,
					Button: hermes.Button{
						Text: option.ButtonText,
						Link: option.ButtonLink,
					},
				},
			},
		},
	}

	switch option.MimeType {
	case MimeTypeHtml:
		return h.GenerateHTML(email)
	}
	return h.GeneratePlainText(email)
}
