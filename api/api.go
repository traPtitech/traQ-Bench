package api

import (
	"context"
	"fmt"
	"github.com/antihax/optional"
	traqApi "github.com/sapphi-red/go-traq"
)

type User struct {
	session string
	client  *traqApi.APIClient
}

func NewUser(id string, pass string) User {
	var user User

	client := traqApi.NewAPIClient(traqApi.NewConfiguration())
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
		fmt.Println("Failed to get session")
		return User{}
	}

	conf := traqApi.NewConfiguration()
	conf.AddDefaultHeader("Session", fmt.Sprintf("r_session=%s", user.session))
	user.client = traqApi.NewAPIClient(conf)

	return user
}

func (user *User) createUser(id string, pass string) User {
	_, err := user.client.UserApi.UsersPost(context.Background(), &traqApi.UsersPostOpts{
		InlineObject4: optional.NewInterface(
			traqApi.InlineObject4{
				Name:     id,
				Password: pass,
			}),
	})
	if err != nil {
		fmt.Printf("Failed to create user for id %s and pass %s: %s\n", id, pass, err)
		return User{}
	}

	fmt.Printf("Successfully created user with id %s", id)
	return NewUser(id, pass)
}
