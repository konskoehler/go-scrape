package sale

import (
	"time"
)

// Sale is one successful auction on ebay.de
type Sale struct {
	Title            string      `json:"title"`
	Date             time.Time   `json:"date"`
	Cost             int         `json:"cost"`
	ProposalAccepted bool        `json:"proposal_accepted"`
	Shipping         int         `json:"shipping"`
	URL              string      `json:"url"`
	Seller           string      `json:"seller"`
	Detail           interface{} `json:"detail"`
}
