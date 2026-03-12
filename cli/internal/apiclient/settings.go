package apiclient

// Setting represents a key-value setting from the API.
type Setting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GetSettings returns all settings.
func (c *Client) GetSettings() ([]Setting, error) {
	var settings []Setting
	err := c.Get("/settings", &settings)
	return settings, err
}

// UpdateSetting updates a single setting by key.
func (c *Client) UpdateSetting(key, value string) error {
	var result Setting
	return c.Put("/settings/"+key, map[string]string{"value": value}, &result)
}
