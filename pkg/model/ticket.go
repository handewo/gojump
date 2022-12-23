package model

type TicketState struct {
	Approver string `json:"approver,omitempty"`
	State    string `json:"state"`
}

const (
	TicketOpen     = "pending"
	TicketApproved = "approved"
	TicketRejected = "rejected"
	TicketClosed   = "closed"
)

type AssetLoginTicketInfo struct {
	TicketId    string   `json:"ticket_id"`
	NeedConfirm bool     `json:"need_confirm"`
	Reviewers   []string `json:"reviewers"`
}

type LoginTicket struct {
	TicketId        string
	State           string
	Approver        string
	ApplicationDate string
	ApproveDate     string
	Username        string
	AssetName       string
	SysUsername     string
}
