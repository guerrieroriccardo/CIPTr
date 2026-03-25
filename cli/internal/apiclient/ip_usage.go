package apiclient

import "github.com/guerrieroriccardo/CIPTr/backend/models"

// GetIPUsage fetches IP utilization data.
// queryParams should include the leading "?" (e.g. "?site_id=3") or be empty for global.
func (c *Client) GetIPUsage(queryParams string) (models.IPUsageResponse, error) {
	var result models.IPUsageResponse
	err := c.Get("/ip-usage"+queryParams, &result)
	return result, err
}
