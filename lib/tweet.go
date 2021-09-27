package twitter

type Tweet struct {
	ID              string `json:"id"`
	Text            string `json:"text"`
	InReplyToUserID string `json:"in_reply_to_user_id"`
}
