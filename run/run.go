package run

import (
	"encoding/json"
	openapi "github.com/sapphi-red/go-traq"
	constant "github.com/traPtitech/traQ-Bench/const"
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

const (
	WaitBlock = 10
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Run(spec int) {
	maxUsers := constant.MaxUsers
	if 0 < spec && spec < constant.MaxUsers {
		maxUsers = spec
	}
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

	log.Printf("Logging in for %v users", maxUsers)
	max := int(math.Ceil(float64(maxUsers) / float64(WaitBlock)))
	mut := sync.Mutex{}
	loggedIn := make([]*api.User, 0)

	for i := 0; i < max; i++ {
		endIndex := (i + 1) * WaitBlock
		if i == max-1 {
			endIndex = maxUsers
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

	// Do benchmark
	admin, err := api.NewUser("traq", "traq")
	if err != nil {
		log.Println("Failed to prepare traq account")
		return
	}

	channels, err := admin.GetChannels()
	if err != nil {
		log.Println("Failed to get channels")
		return
	}

	spam := getChannelId(channels, "spam")
	if spam == "" {
		log.Println("Failed to get spam channel id?")
		return
	}

	sseReceived := 0
	err400 := 0
	err500 := 0
	errUnknown := 0

	log.Printf("Starting benchmark with %v users\n", len(loggedIn))

	wg := sync.WaitGroup{}
	for _, v := range loggedIn {
		wg.Add(1)
		go func(v *api.User) {
			m := runSingle(v, &wg, spam)
			sseReceived += m.sseReceived
			err400 += m.err400
			err500 += m.err500
			errUnknown += m.errUnknown
		}(v)
	}

	wg.Wait()

	log.Printf("Benchmark finished with %v users\n", len(loggedIn))

	log.Println("SSE events received:", sseReceived)
	log.Println("400 Error:", err400)
	log.Println("500 Error:", err500)
	log.Println("Unknown Error:", errUnknown)
}

func getChannelId(channels []openapi.Channel, name string) string {
	for _, v := range channels {
		if v.Name == name && v.ChannelId != "" {
			return v.ChannelId
		}
	}
	return ""
}

func runSingle(user *api.User, wg *sync.WaitGroup, spam string) *metrics {
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)

	sseReceived := int32(0)
	user.ConnectSSE(&sseReceived, &spam)

	end := time.After(45 * time.Second)
	t := time.NewTicker(3 * time.Second)

	err400 := 0
	err500 := 0
	errUnknown := 0

	_, err := user.GetChannelMessages(spam, 20, 0)
	if err != nil {
		log.Println(user.UserId, "error:", err)
	}

loop:
	for {
		select {
		case <-t.C:
			err := user.PostHeartBeat(api.HeartbeatStatuses[rand.Intn(len(api.HeartbeatStatuses))], spam)
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

			if rand.Float64() < 0.01 {
				_, err = user.PostChannelMessage(spam, randStringRunes(32))
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

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
