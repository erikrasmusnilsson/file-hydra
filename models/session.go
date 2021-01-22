package models

// Session models the required data to keep
// track of a given downloading session.
type Session struct {
	ID               string `json:"id"`
	Filename         string `json:"filename"`
	ConnectedClients int    `json:"connectedClients"`
	ExpectedClients  int    `json:"expectedClients"`
}
