package config

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// defaults
const (
	ConfigFile       = "conf/config.yml"
	ConfigType       = ""
	FailureThreshold = 1
)

type Config struct {
	Port                 uint
	Timeout              uint
	Cron                 string
	NotificationMethods  map[string]map[string]interface{}
	Notifications        []string
	NotificationInterval uint
	Checks               []Check
	DataDir              string
	TZ                   string
	FailureThreshold     int
	FailureInterval      int
}

type Check struct {
	Name                 string
	Url                  string
	Response             string
	StatusCode           int
	ContentType          string
	Cron                 string
	Notifications        []string
	NotificationInterval uint
}

type configManager struct {
	config              Config
	getCh               chan Config
	setCh               chan bool
	chOfCh               chan chan Config
	subscribeEventCh            chan bool
	onChangeSubscribers []chan Config
}

var Conf *configManager

func NewConfig() *configManager {
	c := &configManager{
		getCh:    make(chan Config),
		setCh:    make(chan bool),
		chOfCh:    make(chan chan Config),
		subscribeEventCh: make(chan bool),
	}

	c.loadConfig(false)
	go c.eventManager()

	return c
}

func (c *configManager) Get() Config {
	return <-c.getCh
}

func (c *configManager) SubscribeOnChange() chan Config {
	c.subscribeEventCh <- true
	return <-c.chOfCh
}

func (c *configManager) createSubCh() chan Config {
	ch := make(chan Config)
	c.onChangeSubscribers = append(c.onChangeSubscribers, ch)
	return ch
}

func (c *configManager) eventManager() {
	for {
		select {
		case c.getCh <- c.config:
		case <-c.setCh:
			c.loadConfig(true)
		case <-c.subscribeEventCh:
			ch := make(chan Config)
			c.onChangeSubscribers = append(c.onChangeSubscribers, ch)
			c.chOfCh <- ch
		}
	}
}

func (c *configManager) loadConfig(reload bool) {
	if err := viper.ReadInConfig(); err != nil {
		if !reload {
			log.Fatal(err)
		} else {
			log.Println("Configuration error: '", err, "', still using old configuration")
		}
	}

	var conf_tmp Config
	if err := viper.Unmarshal(&conf_tmp); err != nil {
		log.Fatal(err)
	}
	c.config = conf_tmp
	log.Println("Loaded config ", c.config)

	for _, ch := range c.onChangeSubscribers {
		select {
		case ch <- c.config:
		default:
			log.Println("default block")
			break
		}
	}
}

func (c *configManager) changeEventHandler(e fsnotify.Event) {
	log.Printf("Config file %v changed, reloading", e.Name)
	c.setCh <- true
}

func init() {
	viper.AutomaticEnv()
	// defaults for env variables
	viper.SetDefault("CONFIG_FILE", ConfigFile)
	viper.SetDefault("CONFIG_TYPE", ConfigType)
	// default for config file parameters
	viper.SetDefault("failureThreshold", FailureThreshold)

	if config_type := viper.GetString("CONFIG_TYPE"); config_type != "" {
		viper.SetConfigType(config_type)
	}

	viper.SetConfigFile(viper.GetString("CONFIG_FILE"))

	Conf = NewConfig()

	viper.OnConfigChange(Conf.changeEventHandler)
	viper.WatchConfig()
	log.Println("Init finished")
}
