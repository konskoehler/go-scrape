package sale

import (
	"time"
)

// Sale is one successful auction on ebay.de
type Sale struct {
	Title            string      `json:"title"`
	DateSold         time.Time   `json:"date_sold"`
	DateScraped      time.Time   `json:"date_scraped"`
	Cost             int         `json:"cost"`
	ProposalAccepted bool        `json:"proposal_accepted"`
	Shipping         int         `json:"shipping"`
	URL              string      `json:"url"`
	Seller           string      `json:"seller"`
	Detail           interface{} `json:"detail"`
}
