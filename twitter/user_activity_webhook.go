package twitter

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// WebHookRequest is sent from twitter server as webhook payload.
// https://developer.twitter.com/en/docs/accounts-and-users/subscribe-account-activity/guides/account-activity-data-objects
type accountActivityPayload struct {
	ForUserID           *string              `json:"for_user_id"`
	IsBlockedBy         *bool                `json:"is_blocked_by"`
	UserHasBlocked      *bool                `json:"user_has_blocked"`
	Users               map[string]User      `json:"users"`
	TweetCreateEvents   []TweetCreateEvent   `json:"tweet_create_events"`
	TweetDeleteEvents   []DeleteEvent        `json:"tweet_delete_events"`
	FavoriteEvents      []FavoriteEvent      `json:"favorite_events"`
	FollowEvents        []FriendshipEvent    `json:"follow_events"`
	BlockEvents         []FriendshipEvent    `json:"block_events"`
	MuteEvents          []FriendshipEvent    `json:"mute_events"`
	DirectMessageEvents []DirectMessageEvent `json:"direct_message_events"`
	UserEvent           *struct {
		Revoke Revoke `json:"revoke"`
	} `json:"user_event"`
}

type TweetCreateEvent Tweet

type DMEvent struct {
	DirectMessageEvent
	Users map[string]User
}

type DeleteEvent struct {
	Status    Status `json:"status"`
	Timestamp string `json:"timestamp_ms"`
}

type FavoriteEvent Tweet

type FriendshipEvent struct {
	Type             string `json:"type"`
	CreatedTimestamp string `json:"created_timestamp"`
	Target           []User `json:"target"`
	Source           []User `json:"source"`
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

func readRequestBody(reader *io.ReadCloser) string {
	buf, _ := ioutil.ReadAll(*reader)
	s := string(buf)
	*reader = ioutil.NopCloser(bytes.NewBuffer(buf))
	return s
}

func generateCRCToken(crcToken, consumerSecret string) string {
	mac := hmac.New(sha256.New, []byte(consumerSecret))
	mac.Write([]byte(crcToken))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func CreateCRCHandler(consumerSecret string) func(*gin.Context) {
	return func(c *gin.Context) {
		if token, ok := c.GetQuery("crc_token"); ok {
			c.JSON(http.StatusOK, gin.H{
				"response_token": "sha256=" + generateCRCToken(token, consumerSecret),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Missing crc_token in the request.",
			})
		}
	}
}

func CreateTwitterAuthHandler(consumerSecret string) func(*gin.Context) {
	return func(c *gin.Context) {
		signature := c.Request.Header.Get("X-Twitter-Webhooks-Signature")

		if signature == "sha256="+generateCRCToken(readRequestBody(&c.Request.Body), consumerSecret) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "The webhook signature is not correct."})
		}
	}
}

func CreateWebhookHandler(handler chan interface{}) (func(*gin.Context), error) {
	if handler == nil {
		return nil, errors.New("nil account activity handler were passed")
	}
	return func(c *gin.Context) {
		var req accountActivityPayload
		if err := c.BindJSON(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			handler <- fmt.Errorf("an error occurred while parsing json: %+v", err)
			return
		}

		if req.ForUserID == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			done := make(chan interface{})

			go func() {
				defer func() {
					if r := recover(); r != nil {
						handler <- fmt.Errorf("error occurred while handling webhook: %s", r)
					}
				}()

				for _, e := range req.TweetCreateEvents {
					handler <- e
				}
				for _, e := range req.TweetDeleteEvents {
					handler <- e
				}
				for _, e := range req.FavoriteEvents {
					handler <- e
				}
				for _, e := range req.FollowEvents {
					handler <- e
				}
				for _, e := range req.BlockEvents {
					handler <- e
				}
				for _, e := range req.MuteEvents {
					handler <- e
				}
				for _, e := range req.DirectMessageEvents {
					handler <- DMEvent{
						e,
						req.Users,
					}
				}
				if req.UserEvent != nil {
					handler <- req.UserEvent.Revoke
				}
				close(done)
			}()

			select {
			case <-ctx.Done(): // Timeout
				handler <- errors.New("call to the webhook handler timed out")
			case <-done:
				return
			}
		}
	}, nil
}
