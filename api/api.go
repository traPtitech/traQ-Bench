package api

import (
	"context"
	"fmt"
	"github.com/antihax/optional"
	traqApi "github.com/sapphi-red/go-traq"
	"log"
	"sync/atomic"
)

var (
	baseUrl           = "http://localhost:3000/api/1.0"
	HeartbeatStatuses = []HeartBeatStatus{None, Monitoring, Editing}
	SseEvents         = []string{"USER_JOINED", "USER_LEFT", "USER_TAGS_UPDATED", "USER_GROUP_UPDATED", "USER_ICON_UPDATED", "USER_ONLINE", "USER_OFFLINE", "CHANNEL_CREATED", "CHANNEL_DELETED", "CHANNEL_UPDATED", "CHANNEL_STARED", "CHANNEL_UNSTARED", "CHANNEL_VISIBILITY_CHANGED", "MESSAGE_CREATED", "MESSAGE_UPDATED", "MESSAGE_DELETED", "MESSAGE_READ", "MESSAGE_STAMPED", "MESSAGE_UNSTAMPED", "MESSAGE_PINNED", "MESSAGE_UNPINNED", "MESSAGE_CLIPPED", "MESSAGE_UNCLIPPED", "STAMP_CREATED", "STAMP_DELETED", "TRAQ_UPDATED"}
)

type HeartBeatStatus string

const (
	None       HeartBeatStatus = "none"
	Monitoring HeartBeatStatus = "monitoring"
	Editing    HeartBeatStatus = "editing"
)

type User struct {
	UserId   string             `json:"id"`
	Password string             `json:"password"`
	session  string             `json:"-"`
	client   *traqApi.APIClient `json:"-"`
}

func newDevConfiguration() *traqApi.Configuration {
	conf := traqApi.NewConfiguration()
	conf.BasePath = baseUrl
	return conf
}

// ユーザーとしてログインし新しいユーザーインスタンスを返します。
func NewUser(id string, pass string) (*User, error) {
	var user User
	user.UserId = id
	user.Password = pass

	err := user.Login()
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (user *User) Login() error {
	client := traqApi.NewAPIClient(newDevConfiguration())
	res, err := client.AuthenticationApi.LoginPost(context.Background(), &traqApi.LoginPostOpts{
		Redirect: optional.EmptyString(),
		InlineObject: optional.NewInterface(
			traqApi.InlineObject{
				Name: user.UserId,
				Pass: user.Password,
			}),
	})
	if err != nil {
		return err
	}

	for _, v := range res.Cookies() {
		if v.Name == "r_session" {
			user.session = v.Value
		}
	}

	if user.session == "" {
		err := fmt.Errorf("failed to get session")
		return err
	}

	conf := newDevConfiguration()
	conf.AddDefaultHeader("Cookie", fmt.Sprintf("r_session=%s", user.session))
	user.client = traqApi.NewAPIClient(conf)

	log.Printf("Successfully logged in for user %s\n", user.UserId)
	return nil
}

// 新しいユーザーを作成します。
func (user *User) CreateUser(id string, pass string) (*User, error) {
	_, err := user.client.UserApi.UsersPost(context.Background(), &traqApi.UsersPostOpts{
		InlineObject4: optional.NewInterface(
			traqApi.InlineObject4{
				Name:     id,
				Password: pass,
			}),
	})
	if err != nil {
		err := fmt.Errorf("failed to create user for id %s and pass %s: %s\n", id, pass, err)
		return &User{}, err
	}

	log.Printf("Successfully created user with id %s\n", id)
	return NewUser(id, pass)
}

func (user *User) PostHeartBeat(status HeartBeatStatus, channelId string) error {
	_, err := user.client.HeartbeatApi.HeartbeatPost(context.Background(), &traqApi.HeartbeatPostOpts{
		InlineObject14: optional.NewInterface(
			traqApi.InlineObject14{
				Status:    string(status),
				ChannelId: channelId,
			}),
	})
	return err
}

func (user *User) ConnectSSE(sseReceived *int32) {
	// log.Printf("Connecting sse for user %s\n", user.UserId)
	ch, err := OpenURL(user, baseUrl+"/notification")
	if err != nil {
		log.Printf("Failed to connect sse for user %s: %s", user.UserId, err)
		return
	}

	go func() {
		for _ = range ch {
			// log.Printf("%s sse event %s received: %s\n", user.UserId, e.Name, e.Data)
			atomic.AddInt32(sseReceived, 1)
		}
	}()
}

func (user *User) GetChannels() ([]traqApi.Channel, error) {
	channels, _, err := user.client.ChannelApi.ChannelsGet(context.Background())
	if err != nil {
		return nil, err
	}
	return channels, nil
}
