package mail

import (
	"github.com/matcornic/hermes"
)

type EmailContentOption struct {
	Product     Product
	ClientName  string
	Intros      []string
	Message     string
	Instruction string
	DataTable   [][]Data
	Button      Button
	Outros      []string
	MimeType    MimeType
}

type Product struct {
	Name string
	Link string
	Logo string
}

type Button struct {
	Color     string // example: #22BC66
	TextColor string
	Text      string
	Link      string
}

type Data struct {
	Key   string
	Value string
}

func GenerateTransactionalEmail(option EmailContentOption) (string, error) {
	// Configure hermes by setting a theme and your product info
	h := hermes.Hermes{
		// Optional Theme
		// Theme: new(Default)
		Product: hermes.Product{
			// Appears in header & footer of e-mails
			Name: option.Product.Name,
			Link: option.Product.Link,
			// Optional product logo
			Logo: option.Product.Logo,
		},
	}

	var entries []hermes.Entry
	var table [][]hermes.Entry

	for _, rows := range option.DataTable {
		for _, row := range rows {
			entry := hermes.Entry{
				Key:   row.Key,
				Value: row.Value,
			}
			entries = append(entries, entry)
		}

		table = append(table, entries)
		entries = nil
	}

	email := hermes.Email{
		Body: hermes.Body{
			Name:         option.ClientName,
			Intros:       option.Intros,
			FreeMarkdown: hermes.Markdown(option.Message),
			Table: hermes.Table{
				Data: table,
			},
			Actions: []hermes.Action{
				{
					Instructions: option.Instruction,
					Button:       hermes.Button(option.Button),
				},
			},
			Outros: option.Outros,
		},
	}

	switch option.MimeType {
	case MimeTypeHtml:
		emailBody, err := h.GenerateHTML(email)
		if err != nil {
			return "", err
		}

		return emailBody, nil
	default:
		emailText, err := h.GeneratePlainText(email)
		if err != nil {
			return "", err
		}

		return emailText, nil
	}
}
