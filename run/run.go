package run

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/traPtitech/traQ-Bench/api"
)

type AtomicBool struct {
	Bool bool
}

type metrics struct {
	sseReceived int
	err400      int
	err500      int
	errUnknown  int
}

var (
	MaxUsers  = 300
	WaitBlock = 10
)

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

	log.Printf("Logging in for %v users", MaxUsers)
	max := int(math.Ceil(float64(MaxUsers) / float64(WaitBlock)))
	mut := sync.Mutex{}
	loggedIn := make([]*api.User, 0)

	for i := 0; i < max; i++ {
		endIndex := (i + 1) * WaitBlock
		if i == max-1 {
			endIndex = MaxUsers
		}
		usersToLogin := users[i*WaitBlock : endIndex]

		wg := sync.WaitGroup{}
		for _, v := range usersToLogin {
			wg.Add(1)
			go func(user *api.User) {
				err := user.Login()
				if err != nil {
					log.Println("User "+user.UserId+" login failed", err)
					return
				}
				mut.Lock()
				loggedIn = append(loggedIn, user)
				mut.Unlock()
				wg.Done()
			}(v)
		}
		wg.Wait()
	}

	log.Printf("%v users login finished\n", len(loggedIn))
	log.Println("Starting benchmark")

	// Do benchmark
	admin, err := api.NewUser("traq", "traq")
	if err != nil {
		log.Println("Failed to prepare traq account")
		return
	}

	channelId := ""
	channels, err := admin.GetChannels()
	if err != nil {
		log.Println("Failed to get channels")
		return
	}
	for _, v := range channels {
		if v.Name == "general" && v.ChannelId != "" {
			channelId = v.ChannelId
			break
		}
	}
	if channelId == "" {
		log.Println("Couldn't find channel general?")
		return
	}

	sseReceived := 0
	err400 := 0
	err500 := 0
	errUnknown := 0

	wg := sync.WaitGroup{}
	for _, v := range loggedIn {
		wg.Add(1)
		go func(v *api.User) {
			m := runSingle(v, &wg, channelId)
			sseReceived += m.sseReceived
			err400 += m.err400
			err500 += m.err500
			errUnknown += m.errUnknown
		}(v)
	}

	wg.Wait()

	log.Println("Benchmark finished")

	log.Println("SSE events received:", sseReceived)
	log.Println("400 Error:", err400)
	log.Println("500 Error:", err500)
	log.Println("Unknown Error:", errUnknown)
}

func runSingle(user *api.User, wg *sync.WaitGroup, channelId string) *metrics {
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)

	sseReceived := int32(0)
	user.ConnectSSE(&sseReceived)

	end := time.After(45 * time.Second)
	t := time.NewTicker(3 * time.Second)

	err400 := 0
	err500 := 0
	errUnknown := 0

loop:
	for {
		select {
		case <-t.C:
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

	return &metrics{
		sseReceived: int(sseReceived),
		err400:      err400,
		err500:      err500,
		errUnknown:  errUnknown,
	}
}
