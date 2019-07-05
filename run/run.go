package run

import (
	"encoding/json"
	"fmt"
	traqApi "github.com/sapphi-red/go-traq"
	"github.com/traPtitech/traQ-Bench/api"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"
)

type AtomicBool struct {
	Bool bool
}

func Run() {
	fmt.Println("run")

	if _, err := os.Stat("./users.json"); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("users.json does not exist, please run init first.")
			return
		} else {
			fmt.Println("Something went wrong while getting file info", err)
			return
		}
	}

	file, err := os.Open("./users.json")
	if err != nil {
		fmt.Println("Failed to open file", err)
		return
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("Failed to read file", err)
		return
	}

	var users []*api.User
	err = json.Unmarshal(bytes, &users)
	if err != nil {
		fmt.Println("Failed to unmarshal users.json file", err)
		return
	}

	fmt.Println("Loaded users.json")

	for i := 0; i < 10; i++ {
		wg := sync.WaitGroup{}
		for j := 0; j < 30; j++ {
			wg.Add(1)
			user := users[i*30+j]
			go func() {
				err := user.Login()
				if err != nil {
					fmt.Println("User "+user.UserId+" login failed", err)
					return
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}

	fmt.Println("Users login finished")
	fmt.Println("Starting benchmark")

	// Do benchmark
	admin, err := api.NewUser("traq", "traq")
	if err != nil {
		fmt.Println("Failed to prepare traq account")
		return
	}
	channels, err := admin.GetChannels()

	wg := sync.WaitGroup{}
	for _, v := range users {
		wg.Add(1)
		runSingle(v, &wg, &channels)
	}

	wg.Wait()

	fmt.Println("Benchmark finished")
}

func runSingle(user *api.User, wg *sync.WaitGroup, channels *[]traqApi.Channel) {
	go func() {
		user.ConnectSSE()

		rand.Seed(time.Now().UnixNano())
		time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)

		end := time.After(45 * time.Second)
		t := time.NewTicker(3 * time.Second)

	loop:
		for {
			select {
			case <-t.C:
				err := user.PostHeartBeat(api.Monitoring, (*channels)[rand.Intn(len(*channels))].ChannelId)
				if err != nil {
					fmt.Println(user.UserId+" error: ", err)
				}
			case <-end:
				break loop
			}
		}
		t.Stop()
		wg.Done()
	}()
}
