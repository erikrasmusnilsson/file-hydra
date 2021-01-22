package models

// Init models the data to be sent from
// a client upon an initiation request.
type Init struct {
	Filename        string `json:"filename"`
	ExpectedClients int    `json:"expectedClients"`
}
