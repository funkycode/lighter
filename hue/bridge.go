package hue

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"sync"
	"time"

	"github.com/amimof/huego"
)

const (
	bridgeName = "lighter"
)

type Bridge struct {
	*huego.Bridge
	registered bool
}

// Settings contains hue light connection settings
type Settings struct {
	IP   string `json:"ip"`
	User string `json:"user"`
}

func getSettingsFile() (settingsFile string, err error) {
	user, err := user.Current()
	if err != nil {
		return
	}

	settingsDir := path.Join(user.HomeDir, ".config", bridgeName)
	if err = os.MkdirAll(settingsDir, 0755); err != nil {
		return
	}
	settingsFile = path.Join(settingsDir, "settings.json")
	_, err = os.Open(settingsFile)
	if os.IsNotExist(err) {
		_, err = os.Create(settingsFile)
	}
	return
}

// NewSettings loads settings
func NewSettings() (settings Settings, err error) {
	log.Println("loading settings")
	settingsFile, err := getSettingsFile()
	if err != nil {
		return
	}
	file, err := os.Open(settingsFile)
	if err != nil {
		return
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.Decode(&settings)
	return
}

// Save will save settings
func (s *Settings) Save() (err error) {
	settingsFile, err := getSettingsFile()
	if err != nil {
		return
	}

	file, _ := os.OpenFile(settingsFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	defer file.Close()
	decoder := json.NewEncoder(file)
	err = decoder.Encode(&s)
	if err != nil {
		return
	}
	return
}

func (b *Bridge) Connect() {
	settings, err := NewSettings()
	if err != nil {
		return
	}
	b.Bridge = huego.New(settings.IP, settings.User)
	config, err := b.GetConfig()
	fmt.Printf("config: %#v\nerror:%s\n", config, err)
	if err != nil {
		return
	}
	for _, w := range config.Whitelist {
		if w.Username == settings.User {
			b.registered = true
			break
		}
	}
	return
}

func (b *Bridge) IsRegistered() bool {
	return b.registered
}

func (b *Bridge) Register(waitTime time.Duration) (err error) {
	var user string
	var wg sync.WaitGroup
	settings, err := NewSettings()
	if err != nil {
		fmt.Println("Get settings fail:", err)
		return
	}
	bridge, err := huego.Discover()
	if err != nil {
		fmt.Println("Discover fail:", err)
		return
	}
	wg.Add(1)
	go func() {
		for {
			user, err = bridge.CreateUser(bridgeName)
			if err != nil {
				fmt.Println("Failed to register:", err)
				err = nil
				time.Sleep(1 * time.Second)
				continue
			}
			wg.Done()
			break
		}
	}()
	wg.Wait()
	b.Bridge = bridge.Login(user)
	settings.IP = b.Bridge.Host
	settings.User = b.Bridge.User
	err = settings.Save()
	if err != nil {
		fmt.Println("Save fail:", err)
		return
	}
	config, err := b.GetConfig()
	if err != nil {
		fmt.Println("Config:", err)
		return
	}
	for _, w := range config.Whitelist {
		if w.Username == settings.User {
			b.registered = true
			break
		}
	}
	return

}
