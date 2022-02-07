package sale

import (
	"time"
)

// Sale is one successful auction on ebay.de
type Sale struct {
	Title            string
	Date             time.Time
	Cost             int
	ProposalAccepted bool
	Shipping         int
	URL              string
	Seller           string
	Detail           interface{}
}
