package twitter

type Meta struct {
	Sent    string  `json:"sent"`
	Summary summary `json:"summary,omitempty"`
}

type summary struct {
	Created    int `json:"created"`
	NotCreated int `json:"not_created"`
}
