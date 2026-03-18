package resource

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

// generateShortCode produces a 3-character uppercase code from a client name.
//
// Rules:
//   - Name is 3 chars: use as-is
//   - 1 word: first 3 consonants (BERPA → BRP)
//   - 2 words: first 2 consonants of word 1 + first char of word 2 (BRAND-EAT → BRE)
//   - 3+ words: first char of each of the first 3 words (Officine Meccaniche Pontina → OMP)
func generateShortCode(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	if len([]rune(name)) == 3 {
		return strings.ToUpper(name)
	}

	// Split on spaces and hyphens.
	words := strings.FieldsFunc(name, func(r rune) bool {
		return unicode.IsSpace(r) || r == '-'
	})

	switch len(words) {
	case 1:
		return strings.ToUpper(firstConsonants(words[0], 3))
	case 2:
		code := firstConsonants(words[0], 2) + string([]rune(words[1])[0])
		return strings.ToUpper(code)
	default:
		code := string([]rune(words[0])[0]) + string([]rune(words[1])[0]) + string([]rune(words[2])[0])
		return strings.ToUpper(code)
	}
}

func firstConsonants(word string, n int) string {
	var result []rune
	for _, r := range strings.ToLower(word) {
		if !strings.ContainsRune("aeiou", r) && unicode.IsLetter(r) {
			result = append(result, r)
			if len(result) == n {
				break
			}
		}
	}
	// Fall back to first characters if not enough consonants.
	if len(result) < n {
		for _, r := range word {
			if unicode.IsLetter(r) && !containsRune(result, unicode.ToLower(r)) {
				result = append(result, unicode.ToLower(r))
				if len(result) == n {
					break
				}
			}
		}
	}
	// Last resort: just take first n characters.
	if len(result) < n {
		runes := []rune(word)
		for i := 0; i < n && i < len(runes); i++ {
			if !containsRune(result, unicode.ToLower(runes[i])) {
				result = append(result, unicode.ToLower(runes[i]))
			}
		}
	}
	return string(result)
}

func containsRune(s []rune, r rune) bool {
	for _, v := range s {
		if v == r {
			return true
		}
	}
	return false
}

func init() {
	Register("clients", &Def{
		Name:    "Client",
		Plural:  "Clients",
		APIPath: "/clients",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Name", Width: 30},
			{Title: "Code", Width: 10},
			{Title: "Notes", Width: 30},
		},
		ToRow: func(raw any) table.Row {
			c := raw.(*models.Client)
			notes := ""
			if c.Notes != nil {
				notes = *c.Notes
			}
			return table.Row{
				fmt.Sprintf("%d", c.ID),
				c.Name,
				c.ShortCode,
				notes,
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.Client).ID)
		},

		Fields: []Field{
			{Key: "name", Label: "Name", Required: true},
			{Key: "short_code", Label: "Short Code", Required: true},
			{Key: "domain", Label: "Domain (e.g. client.tld)"},
			{Key: "notes", Label: "Notes"},
		},

		DeriveField: func(key, value string) map[string]string {
			if key == "name" {
				return map[string]string{"short_code": generateShortCode(value)}
			}
			return nil
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var clients []models.Client
			if err := client.Get("/clients", &clients); err != nil {
				return nil, err
			}
			result := make([]any, len(clients))
			for i := range clients {
				result[i] = &clients[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.ClientInput{
				Name:      data["name"],
				ShortCode: data["short_code"],
				Domain:    strPtr(data["domain"]),
			}
			if v, ok := data["notes"]; ok && v != "" {
				input.Notes = &v
			}
			var created models.Client
			err := client.Post("/clients", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.ClientInput{
				Name:      data["name"],
				ShortCode: data["short_code"],
				Domain:    strPtr(data["domain"]),
			}
			if v, ok := data["notes"]; ok && v != "" {
				input.Notes = &v
			}
			var updated models.Client
			err := client.Put("/clients/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/clients/" + id)
		},
	})
}
