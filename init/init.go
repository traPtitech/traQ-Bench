package init

import (
	"fmt"
	"github.com/traPtitech/traQ-Bench/api"
)

func Init() {
	fmt.Println("init")

	admin, err := api.NewUser("traq", "traq")
	if err != nil {
		panic(err)
	}

	u, err := admin.CreateUser("testuser01", "testpass01")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(u)
}
