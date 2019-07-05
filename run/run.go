package run

import (
	"fmt"
	"github.com/traPtitech/traQ-Bench/api"
	"strconv"
	"sync"
)

func Run() {
	fmt.Println("run")

	mut := sync.Mutex{}
	users := make([]*api.User, 0)

	for i := 0; i < 30; i++ {
		wg := sync.WaitGroup{}
		for j := 0; j < 10; j++ {
			id := "user" + strconv.Itoa(i*10+j+1)
			pass := "userpassword" + strconv.Itoa(i*10+j+1)

			wg.Add(1)
			go func() {
				user, err := api.NewUser(id, pass)
				if err != nil {
					fmt.Println(err)
					return
				}
				mut.Lock()
				users = append(users, user)
				mut.Unlock()

				wg.Done()
			}()
		}
		wg.Wait()
	}

	// Do benchmark
}
