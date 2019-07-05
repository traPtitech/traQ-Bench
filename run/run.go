package run

import (
	"encoding/json"
	traqApi "github.com/sapphi-red/go-traq"
	"github.com/traPtitech/traQ-Bench/api"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

type AtomicBool struct {
	Bool bool
}

func Run() {
	log.Println("run")

	if _, err := os.Stat("./users.json"); err != nil {
		if os.IsNotExist(err) {
			log.Println("users.json does not exist, please run init first.")
			return
		} else {
			log.Println("Something went wrong while getting file info", err)
			return
		}
	}

	file, err := os.Open("./users.json")
	if err != nil {
		log.Println("Failed to open file", err)
		return
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("Failed to read file", err)
		return
	}

	var users []*api.User
	err = json.Unmarshal(bytes, &users)
	if err != nil {
		log.Println("Failed to unmarshal users.json file", err)
		return
	}

	log.Println("Loaded users.json")

	for i := 0; i < 30; i++ {
		wg := sync.WaitGroup{}
		for j := 0; j < 10; j++ {
			wg.Add(1)
			user := users[i*10+j]
			go func() {
				err := user.Login()
				if err != nil {
					log.Println("User "+user.UserId+" login failed", err)
					return
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}

	log.Println("Users login finished")
	log.Println("Starting benchmark")

	// Do benchmark
	admin, err := api.NewUser("traq", "traq")
	if err != nil {
		log.Println("Failed to prepare traq account")
		return
	}
	channels, err := admin.GetChannels()

	wg := sync.WaitGroup{}
	for _, v := range users {
		wg.Add(1)
		runSingle(v, &wg, &channels)
	}

	wg.Wait()

	log.Println("Benchmark finished")
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
					log.Println(user.UserId+" error: ", err)
				}
			case <-end:
				break loop
			}
		}
		t.Stop()
		wg.Done()
	}()
}
