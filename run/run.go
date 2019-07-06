package run

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	traqApi "github.com/sapphi-red/go-traq"
	"github.com/traPtitech/traQ-Bench/api"
)

type AtomicBool struct {
	Bool bool
}

type httpErrors struct {
	err400     int
	err500     int
	errUnknown int
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

	err400 := 0
	err500 := 0
	errUnknown := 0

	wg := sync.WaitGroup{}
	for _, v := range users {
		wg.Add(1)
		go func(v *api.User) {
			errs := runSingle(v, &wg, &channels)
			err400 += errs.err400
			err500 += errs.err500
			errUnknown += errs.errUnknown
		}(v)
	}

	wg.Wait()

	log.Println("Benchmark finished")

	log.Println("400 Error:", err400)
	log.Println("500 Error:", err500)
	log.Println("Unknown Error:", errUnknown)
}

func runSingle(user *api.User, wg *sync.WaitGroup, channels *[]traqApi.Channel) *httpErrors {
	user.ConnectSSE()

	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)

	end := time.After(45 * time.Second)
	t := time.NewTicker(3 * time.Second)

	err400 := 0
	err500 := 0
	errUnknown := 0

loop:
	for {
		select {
		case <-t.C:
			var channelId string
			for _, v := range *channels {
				if v.Name == "general" && v.ChannelId != "" {
					channelId = v.ChannelId
					break
				}
			}
			if channelId == "" {
				log.Println("Couldn't find channel general?")
				continue
			}

			err := user.PostHeartBeat(api.HeartbeatStatuses[rand.Intn(len(api.HeartbeatStatuses))], channelId)
			if err != nil {
				errStr := err.Error()
				log.Println(user.UserId, "error:", errStr)

				if errStr == "400 Bad Request" {
					err400++
				} else if errStr == "500 Internal Server Error" {
					err500++
				} else {
					errUnknown++
				}
			}
		case <-end:
			break loop
		}
	}

	t.Stop()
	wg.Done()

	return &httpErrors{
		err400:     err400,
		err500:     err500,
		errUnknown: errUnknown,
	}
}
