package twitter

type Rule struct {
	ID    string `json:"id,omitempty"`
	Value string `json:"value"`
	Tag   string `json:"tag,omitempty"`
}
