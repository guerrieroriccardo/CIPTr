package models

// IPUsageNode represents a single node in the IP usage tree.
// Type can be "client", "site", "address_block", "vlan", or "ip".
type IPUsageNode struct {
	ID       int64          `json:"id"`
	Label    string         `json:"label"`
	Type     string         `json:"type"`
	TotalIPs int            `json:"total_ips,omitempty"`
	UsedIPs  int            `json:"used_ips,omitempty"`
	Children []IPUsageNode  `json:"children,omitempty"`
}

// IPUsageResponse wraps the top-level response for the ip-usage endpoint.
type IPUsageResponse struct {
	Level string        `json:"level"`
	Items []IPUsageNode `json:"items"`
}
