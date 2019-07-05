package api

import (
	"context"
	"fmt"
	"github.com/antihax/optional"
	traqApi "github.com/sapphi-red/go-traq"
)

var (
	baseUrl = "http://localhost:3000/api/1.0"
)

type User struct {
	session string
	client  *traqApi.APIClient
}

func newDevConfiguration() *traqApi.Configuration {
	conf := traqApi.NewConfiguration()
	conf.BasePath = baseUrl
	return conf
}

// ユーザーとしてログインし新しいユーザーインスタンスを返します。
func NewUser(id string, pass string) (*User, error) {
	var user User

	client := traqApi.NewAPIClient(newDevConfiguration())
	res, err := client.AuthenticationApi.LoginPost(context.Background(), &traqApi.LoginPostOpts{
		Redirect: optional.EmptyString(),
		InlineObject: optional.NewInterface(
			traqApi.InlineObject{
				Name: id,
				Pass: pass,
			}),
	})
	if err != nil {
		panic(err)
	}

	for _, v := range res.Cookies() {
		if v.Name == "r_session" {
			user = User{
				session: v.Value,
			}
		}
	}

	if user.session == "" {
		err := fmt.Errorf("failed to get session")
		return &User{}, err
	}

	conf := newDevConfiguration()
	conf.AddDefaultHeader("Cookie", fmt.Sprintf("r_session=%s", user.session))
	user.client = traqApi.NewAPIClient(conf)

	return &user, nil
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
