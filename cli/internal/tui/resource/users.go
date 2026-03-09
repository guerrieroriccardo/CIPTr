package resource

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

type user struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func init() {
	Register("users", &Def{
		Name:    "User",
		Plural:  "Users",
		APIPath: "/users",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Username", Width: 20},
			{Title: "Role", Width: 14},
			{Title: "Created", Width: 20},
		},
		ToRow: func(raw any) table.Row {
			u := raw.(*user)
			return table.Row{
				fmt.Sprintf("%d", u.ID),
				u.Username,
				u.Role,
				u.CreatedAt.Local().Format("2006-01-02 15:04:05"),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*user).ID)
		},

		Fields: []Field{
			{Key: "username", Label: "Username", Required: true},
			{Key: "password", Label: "Password"},
			{Key: "role", Label: "Role", Required: true, PickerOptions: []string{"admin", "technician", "viewer", "guest"}},
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var users []user
			if err := client.Get("/users", &users); err != nil {
				return nil, err
			}
			result := make([]any, len(users))
			for i := range users {
				result[i] = &users[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			body := map[string]string{
				"username": data["username"],
				"role":     data["role"],
			}
			if data["password"] != "" {
				body["password"] = data["password"]
			}
			var created user
			err := client.Post("/register", body, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			body := map[string]string{
				"username": data["username"],
				"role":     data["role"],
			}
			if data["password"] != "" {
				body["password"] = data["password"]
			}
			var updated user
			err := client.Put("/users/"+id, body, &updated)
			return &updated, err
		},
	})
}
