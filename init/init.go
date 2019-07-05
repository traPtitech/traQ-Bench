package init

import (
	"fmt"
	"github.com/traPtitech/traQ-Bench/api"
	"strconv"
	"sync"
)

func Init() {
	fmt.Println("init")

	admin, err := api.NewUser("traq", "traq")
	if err != nil {
		panic(err)
	}

	for i := 0; i < 30; i++ {
		wg := sync.WaitGroup{}
		for j := 0; j < 10; j++ {
			id := "user" + strconv.Itoa(i*10+j+1)
			pass := "userpassword" + strconv.Itoa(i*10+j+1)

			wg.Add(1)
			go func() {
				_, err := admin.CreateUser(id, pass)
				if err != nil {
					fmt.Println(err)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}
