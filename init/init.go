package init

import (
	"encoding/json"
	"github.com/traPtitech/traQ-Bench/api"
	"log"
	"os"
	"strconv"
	"sync"
)

func Init() {
	log.Println("init")

	admin, err := api.NewUser("traq", "traq")
	if err != nil {
		panic(err)
	}

	mut := sync.Mutex{}
	users := make([]*api.User, 0)

	for i := 0; i < 30; i++ {
		wg := sync.WaitGroup{}
		for j := 0; j < 10; j++ {
			id := "user" + strconv.Itoa(i*10+j+1)
			pass := "userpassword" + strconv.Itoa(i*10+j+1)

			wg.Add(1)
			go func() {
				user, err := admin.CreateUser(id, pass)
				if err != nil {
					log.Println(err)
				}
				mut.Lock()
				users = append(users, user)
				mut.Unlock()

				wg.Done()
			}()
		}
		wg.Wait()
	}

	bytes, err := json.Marshal(users)
	if _, err := os.Stat("./users.json"); err == nil {
		err = os.Remove("./users.json")
		log.Println("Failed to remove file", err)
	} else if !os.IsNotExist(err) {
		panic(err)
	}

	file, err := os.Create("./users.json")
	if err != nil {
		log.Println("Failed to create to file", err)
	}

	_, err = file.Write(bytes)
	if err != nil {
		log.Println("Failed to output to file", err)
	}

	_ = file.Close()
}

// DumpUsers User情報をusers.jsonに書き出します 既に(user1, userpassword1), (user2, userpassword2)...がローカルにある場合
func DumpUsers() {
	users := make([]*api.User, 0)
	for i := 0; i < 300; i++ {
		users = append(users, &api.User{
			UserId:   "user" + strconv.Itoa(i+1),
			Password: "userpassword" + strconv.Itoa(i+1),
		})
	}

	bytes, err := json.Marshal(users)
	if _, err := os.Stat("./users.json"); err == nil {
		err = os.Remove("./users.json")
		log.Println("Failed to remove file", err)
	} else if !os.IsNotExist(err) {
		panic(err)
	}

	file, err := os.Create("./users.json")
	if err != nil {
		log.Println("Failed to create to file", err)
	}

	_, err = file.Write(bytes)
	if err != nil {
		log.Println("Failed to output to file", err)
	}

	_ = file.Close()
}
