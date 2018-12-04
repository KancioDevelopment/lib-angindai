package mail

import (
	"fmt"
	"time"

	"github.com/matcornic/hermes"
)

type EmailContentAnnouncement struct {
	MimeType     MimeType
	Organization string
	Intros       []string
	Greeting     string
	Signature    string
}

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
