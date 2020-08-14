package main

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

var lightUser string

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

	settingsDir := path.Join(user.HomeDir, ".config", "ligher")
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

func connect() (bridge *huego.Bridge) {
	settings, err := NewSettings()
	if err != nil || settings.User == "" {
		bridge, _ = huego.Discover()
		fmt.Println("Press connect button on bridge (waiting for 10 seconds)")
		time.Sleep(10 * time.Second)
		user, err := bridge.CreateUser("lighter")
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
	} else {
		bridge = huego.New(settings.IP, settings.User)
	}
	return
}

func main() {
	bridge := connect()
	lights, err := bridge.GetLights()
	if err != nil {
		log.Panicf("Failed to get lights: %s\n", err)
	}
	log.Printf("Lights:\n%+v\n", lights)
}
