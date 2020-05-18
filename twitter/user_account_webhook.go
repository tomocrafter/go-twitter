package twitter

// WebHookRequest is sent from twitter server as webhook payload.
// https://developer.twitter.com/en/docs/accounts-and-users/subscribe-account-activity/guides/account-activity-data-objects
type AccountActivityPayload struct {
	ForUserID           *string              `json:"for_user_id"`
	IsBlockedBy         *bool                `json:"is_blocked_by"`
	UserHasBlocked      *bool                `json:"user_has_blocked"`
	Users               map[string]User      `json:"users"`
	TweetCreateEvents   []Tweet              `json:"tweet_create_events"`
	TweetDeleteEvents   []DeleteEvent        `json:"tweet_delete_events"`
	FavoriteEvents      []Tweet              `json:"favorite_events"`
	FollowEvents        []FriendshipEvent    `json:"follow_events"`
	BlockEvents         []FriendshipEvent    `json:"block_events"`
	MuteEvents          []FriendshipEvent    `json:"mute_events"`
	DirectMessageEvents []DirectMessageEvent `json:"direct_message_events"`
	UserEvent           struct {
		Revoke Revoke `json:"revoke"`
	} `json:"user_event"`
}

type FriendshipEvent struct {
	Type             string `json:"type"`
	CreatedTimestamp string `json:"created_timestamp"`
	Target           []User `json:"target"`
	Source           []User `json:"source"`
}

type DeleteEvent struct {
	Status    Status `json:"status"`
	Timestamp string `json:"timestamp_ms"`
}

type Status struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

type Revoke struct {
	DateTime string `json:"date_time"`
	Target   struct {
		AppID string `json:"app_id"`
	} `json:"target"`
	Source struct {
		UserID string `json:"user_id"`
	} `json:"source"`
}
