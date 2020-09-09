package hue

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"time"

	"github.com/amimof/huego"
)

const (
	bridgeName = "lighter"
)

type Bridge struct {
	*huego.Bridge
	registered bool
	connected  bool
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
	err = decoder.Decode(&settings)
	if err != nil {
		return
	}
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

func Connect() (bridge *huego.Bridge) {
	settings, _ := NewSettings()
	bridge = huego.New(settings.IP, settings.User)
	return
}

func Register() {
	settings, err := NewSettings()
	bridge, _ := huego.Discover()
	fmt.Println("Press connect button on bridge (waiting for 10 seconds)")
	time.Sleep(10 * time.Second)
	user, err := bridge.CreateUser(bridgeName)
	if err != nil {
		log.Panicf("Failed to register user: %s\n", err)
	}
	bridge = bridge.Login(user)
	settings.IP = bridge.Host
	settings.User = bridge.User
	err = settings.Save()
	if err != nil {
		log.Panicf("Failed to save settings: %s\n", err)
	}
}
