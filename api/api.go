package api

import (
	"context"
	"fmt"
	"github.com/antihax/optional"
	"github.com/r3labs/sse"
	traqApi "github.com/sapphi-red/go-traq"
)

var (
	baseUrl = "http://localhost:3000/api/1.0"
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
		panic(err)
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

	fmt.Printf("Successfully created user with id %s\n", id)
	return NewUser(id, pass)
}

type HeartBeatStatus string

const (
	None       HeartBeatStatus = "none"
	Monitoring HeartBeatStatus = "monitoring"
	Editing    HeartBeatStatus = "editing"
)

func (user *User) PostHeartBeat(status HeartBeatStatus, channelId string) error {
	_, err := user.client.HeartbeatApi.HeartbeatPost(context.Background(), string(status), channelId)
	return err
}

func (user *User) ConnectSSE() {
	_ = sse.NewClient(baseUrl + "/notification")
}

func (user *User) GetChannels() ([]traqApi.Channel, error) {
	channels, _, err := user.client.ChannelApi.ChannelsGet(context.Background())
	if err != nil {
		return nil, err
	}
	return channels, nil
}
