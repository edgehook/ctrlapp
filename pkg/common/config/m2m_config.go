package config

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/edgehook/ctrlapp/pkg/common/crypto"
	"github.com/edgehook/ctrlapp/pkg/utils"
	"github.com/fsnotify/fsnotify"
	"k8s.io/klog/v2"
)

const (
	DEFAULT_BROKER_USERNAME = "admin"
	DEFAULT_BROKER_PASSWD   = "vrG8aUHDU9a0pk5b8DEunA=="
)

var onceM2m sync.Once
var LWM2M_CONFIG *Lwm2mYamlConfig

type Lwm2mYamlConfig struct {
	configYamlPath string
	m2mConfig      *Config
	// event callback.
	callback    func(ev *Event)
	configCache *TransportConfig
}

// TransportConfig
type TransportConfig struct {
	Broker       string
	User, Passwd string
	ClientID     string
	DeviceName   string
	// QOS indicates mqtt qos
	// 0: QOSAtMostOnce, 1: QOSAtLeastOnce, 2: QOSExactlyOnce
	// default 0
	// Note: Can not use "omitempty" option,  It will affect the output of the default configuration file
	QOS          int
	CaFilePath   string
	CertFilePath string
	KeyFilePath  string
	ServerEPID   string
	ServerAppID  string
	ClientAppID  string
	OrgID        string
}

func (t *TransportConfig) IsNeedSwitch(o *TransportConfig) bool {
	return reflect.DeepEqual(*t, *o) != true
}

func NewLwm2mYamlConfig(yamlPath string) *Lwm2mYamlConfig {
	configDir := filepath.Dir(yamlPath)
	configYaml := filepath.Base(yamlPath)
	yamlConfig := NewYamlConfig(configDir, configYaml)

	yConfig := &Lwm2mYamlConfig{
		configYamlPath: yamlPath,
		m2mConfig:      yamlConfig,
	}

	//cache the config.
	yConfig.configCache = yConfig.getTransportConfig()

	return yConfig
}

func (lyc *Lwm2mYamlConfig) getTransportConfig() *TransportConfig {
	transConfig := &TransportConfig{}

	transConfig.Broker = lyc.m2mConfig.GetString("transport.mqtt.broker")
	transConfig.User = lyc.m2mConfig.GetString("transport.mqtt.usr")
	transConfig.Passwd = lyc.m2mConfig.GetString("transport.mqtt.passwd")
	transConfig.ClientID = lyc.m2mConfig.GetString("transport.mqtt.clientid")
	transConfig.QOS = lyc.m2mConfig.GetInt("transport.mqtt.qos")
	transConfig.CaFilePath = lyc.m2mConfig.GetString("transport.mqtt.cafile")
	transConfig.CertFilePath = lyc.m2mConfig.GetString("transport.mqtt.certfile")
	transConfig.KeyFilePath = lyc.m2mConfig.GetString("transport.mqtt.keyfile")

	transConfig.DeviceName = lyc.m2mConfig.GetString("m2mclient.device-name")
	transConfig.ServerEPID = lyc.m2mConfig.GetString("m2mclient.server-endpoint-id")
	transConfig.ServerAppID = lyc.m2mConfig.GetString("m2mclient.server-application-id")
	transConfig.ClientAppID = lyc.m2mConfig.GetString("m2mclient.client-application-id")
	transConfig.OrgID = lyc.m2mConfig.GetString("m2mclient.orgid")

	lyc.configCache = transConfig
	return transConfig
}

func GetTransportConfig() *TransportConfig {
	return LWM2M_CONFIG.getTransportConfig()
}

func (lyc *Lwm2mYamlConfig) saveTransportConfig(config *TransportConfig) error {
	if config == nil {
		return errors.New("config is nil")
	}

	defer func() {
		//we must recreate the viper since viper will override the v.ovveride
		// and let v.Get always use old value.
		yamlConfig := NewLwm2mYamlConfig(lyc.configYamlPath)
		lyc.m2mConfig = yamlConfig.m2mConfig

		lyc.configCache = lyc.getTransportConfig()
		if lyc.callback != nil {
			lyc.AddWatchConfig(lyc.callback)
		}
	}()

	if config.Broker != "" {
		lyc.m2mConfig.SetString("transport.mqtt.broker", config.Broker)
	}

	if config.User != "" {
		lyc.m2mConfig.SetString("transport.mqtt.usr", config.User)
	}

	if config.Passwd != "" {
		lyc.m2mConfig.SetString("transport.mqtt.passwd", config.Passwd)
	}

	if config.ClientID != "" {
		lyc.m2mConfig.SetString("transport.mqtt.clientid", config.ClientID)
	}

	if config.QOS >= 0 {
		lyc.m2mConfig.SetInt("transport.mqtt.qos", config.QOS)
	}

	if config.CaFilePath != "" {
		lyc.m2mConfig.SetString("transport.mqtt.cafile", config.CaFilePath)
	}

	if config.CertFilePath != "" {
		lyc.m2mConfig.SetString("transport.mqtt.certfile", config.CertFilePath)
	}

	if config.KeyFilePath != "" {
		lyc.m2mConfig.SetString("transport.mqtt.keyfile", config.KeyFilePath)
	}

	if config.DeviceName != "" {
		lyc.m2mConfig.SetString("m2mclient.device-name", config.DeviceName)
	}

	if config.ServerEPID != "" {
		lyc.m2mConfig.SetString("m2mclient.server-endpoint-id", config.ServerEPID)
	}

	if config.ServerAppID != "" {
		lyc.m2mConfig.SetString("m2mclient.server-application-id", config.ServerAppID)
	}

	if config.ClientAppID != "" {
		lyc.m2mConfig.SetString("m2mclient.client-application-id", config.ClientAppID)
	}

	if config.OrgID != "" {
		lyc.m2mConfig.SetString("m2mclient.orgid", config.OrgID)
	}

	//save config to conf/lwm2m.yaml
	return lyc.m2mConfig.SaveConfig()
}

func SaveTransportConfig(config *TransportConfig) error {
	//save config to conf/lwm2m.yaml
	return LWM2M_CONFIG.saveTransportConfig(config)
}

// SetTransportConfig
func (lyc *Lwm2mYamlConfig) setTransportConfig(config string) error {
	transportConfig := GetTransportConfig()

	configJson := make(map[string]interface{})
	err := json.Unmarshal([]byte(config), &configJson)
	if err != nil {
		klog.Errorf("err: %v", err)
		return err
	}

	if url, ok := configJson["url"]; ok {
		baseUrl := url.(string)

		services, ok := configJson["services"]
		if !ok {
			return nil
		}

		mservices, isThisType := services.([]interface{})
		if !isThisType {
			return errors.New("No such service type")
		}

		var mqttKey = ""
		for _, iv := range mservices {
			service, isThisType := iv.(map[string]interface{})
			if !isThisType {
				return errors.New("iv Not map[string]interface{} type")
			}

			value, ok := service["name"]
			if ok && value.(string) == "mqtt" {
				v, ok := service["key"]
				if ok {
					mqttKey = v.(string)
				}
				break
			}
		}

		if mqttKey == "" {
			return errors.New("mqttKey is empty")
		}

		mqttUrl := baseUrl + mqttKey

		result, err := utils.HttpGet(mqttUrl)
		if err != nil {
			klog.Errorf("err: %v", err)
			return err
		}

		if result == "" {
			return errors.New("No token")
		}

		jsonToken := make(map[string]interface{})
		err = json.Unmarshal([]byte(result), &jsonToken)
		if err != nil {
			klog.Errorf("err: %v", err)
			return err
		}

		hostUrl, ok := jsonToken["serviceHost"]
		if !ok {
			return errors.New("No serviceHost in token")
		}
		host := hostUrl.(string)

		credential, ok := jsonToken["credential"]
		if !ok {
			return errors.New("No credential in token")
		}

		credentialObj, isThisType := credential.(map[string]interface{})
		if !isThisType {
			return errors.New("json invalid data")
		}

		protocols, ok := credentialObj["protocols"]
		if !ok {
			return errors.New("No protocols in credential")
		}

		protocolsObj, isThisType := protocols.(map[string]interface{})
		if !isThisType {
			return errors.New("json invalid data")
		}

		mqtt, ok := protocolsObj["mqtt"]
		if !ok {
			return errors.New("No mqtt in protocols")
		}

		mqttObj, isThisType := mqtt.(map[string]interface{})
		if !isThisType {
			return errors.New("json invalid data")
		}

		v, ok := mqttObj["username"]
		if !ok {
			return errors.New("No username in mqtt")
		}
		userName := v.(string)

		v, ok = mqttObj["password"]
		if !ok {
			return errors.New("No password in mqtt")
		}
		password := v.(string)

		encryptedPwd, err := crypto.Encrypt([]byte(password))
		if err != nil {
			klog.Errorf("err: %v", err)
			return err
		}

		//update the mqtt config
		transportConfig.Broker = host
		transportConfig.User = userName
		transportConfig.Passwd = encryptedPwd
	} else {
		//klog.Infof("urlPath: %v", urlPath)
		var ipAddr = ""
		services, ok := configJson["services"]
		if !ok {
			return nil
		}

		mservices, isThisType := services.([]interface{})
		if !isThisType {
			return errors.New("No such service type")
		}

		for _, iv := range mservices {
			service, isThisType := iv.(map[string]interface{})
			if !isThisType {
				return errors.New("iv Not map[string]interface{} type")
			}

			v, ok := service["name"]
			value := v.(string)
			if ok && value == "mqtt" {
				if _, ok = service["path"]; ok {

				} else {
					klog.Infof("old scan mode")
					v, ok = service["ip"]
					if ok {
						ipAddr = v.(string)
					}

					if ipAddr == "" {
						return errors.New("No ip address")
					}

					//update config
					transportConfig.Broker = ipAddr
					transportConfig.User = DEFAULT_BROKER_USERNAME
					transportConfig.Passwd = DEFAULT_BROKER_PASSWD
				}
				break
			}
		}
	}

	return lyc.saveTransportConfig(transportConfig)
}

func SetTransportConfig(config string) error {
	return LWM2M_CONFIG.setTransportConfig(config)
}

// set device name.
func SetDeviceName(deviceName string) error {
	if deviceName == "" {
		return nil
	}

	transportConfig := GetTransportConfig()
	transportConfig.DeviceName = deviceName
	return SaveTransportConfig(transportConfig)
}

func LoadM2mConfig(configYamlPath string) {
	onceM2m.Do(func() {
		LWM2M_CONFIG = NewLwm2mYamlConfig(configYamlPath)
	})
}

// Event generated when any config changes
type Event struct {
	EventSource string
	Key         string
	Value       interface{}
	HasUpdated  bool
}

func (lyc *Lwm2mYamlConfig) AddWatchConfig(callback func(ev *Event)) {
	lyc.callback = callback
	lyc.m2mConfig.WatchConfig()
	lyc.m2mConfig.OnConfigChange(func(event fsnotify.Event) {
		switch event.Op {
		case fsnotify.Create:
		case fsnotify.Write:
			time.Sleep(time.Millisecond * 100)
			events := lyc.compareUpdate(event.Name)
			for _, ev := range events {
				callback(ev)
			}
		}
	})
}

func AddWatchConfig(callback func(ev *Event)) {
	LWM2M_CONFIG.AddWatchConfig(callback)
}

func (lyc *Lwm2mYamlConfig) compareUpdate(fname string) []*Event {
	events := make([]*Event, 0)

	oldTransConfig := lyc.configCache
	//we must reload the config since OnConfigChange call reload config
	// too early!
	lyc.m2mConfig.ReloadConfig()
	newtC := lyc.getTransportConfig()

	if newtC.Broker != "" && newtC.Broker != oldTransConfig.Broker {
		e := &Event{
			EventSource: fname,
			Key:         "transport.mqtt.broker",
			Value:       newtC.Broker,
			HasUpdated:  true,
		}
		events = append(events, e)
	}

	if newtC.User != "" && newtC.User != oldTransConfig.User {
		e := &Event{
			EventSource: fname,
			Key:         "transport.mqtt.usr",
			Value:       newtC.User,
			HasUpdated:  true,
		}
		events = append(events, e)
	}

	if newtC.Passwd != "" && newtC.Passwd != oldTransConfig.Passwd {
		e := &Event{
			EventSource: fname,
			Key:         "transport.mqtt.passwd",
			Value:       newtC.Passwd,
			HasUpdated:  true,
		}
		events = append(events, e)
	}

	if newtC.ClientID != "" && newtC.ClientID != oldTransConfig.ClientID {
		e := &Event{
			EventSource: fname,
			Key:         "transport.mqtt.clientid",
			Value:       newtC.ClientID,
			HasUpdated:  true,
		}
		events = append(events, e)
	}

	if newtC.QOS != oldTransConfig.QOS {
		e := &Event{
			EventSource: fname,
			Key:         "transport.mqtt.qos",
			Value:       newtC.QOS,
			HasUpdated:  true,
		}
		events = append(events, e)
	}

	if newtC.DeviceName != "" && newtC.DeviceName != oldTransConfig.DeviceName {
		e := &Event{
			EventSource: fname,
			Key:         "m2mclient.device-name",
			Value:       newtC.DeviceName,
			HasUpdated:  true,
		}
		events = append(events, e)
	}

	if newtC.ServerEPID != "" && newtC.ServerEPID != oldTransConfig.ServerEPID {
		e := &Event{
			EventSource: fname,
			Key:         "m2mclient.server-endpoint-id",
			Value:       newtC.ServerEPID,
			HasUpdated:  true,
		}
		events = append(events, e)
	}

	if newtC.ServerAppID != "" && newtC.ServerAppID != oldTransConfig.ServerAppID {
		e := &Event{
			EventSource: fname,
			Key:         "m2mclient.server-application-id",
			Value:       newtC.ServerAppID,
			HasUpdated:  true,
		}
		events = append(events, e)
	}

	if newtC.ClientAppID != "" && newtC.ClientAppID != oldTransConfig.ClientAppID {
		e := &Event{
			EventSource: fname,
			Key:         "m2mclient.client-application-id",
			Value:       newtC.ClientAppID,
			HasUpdated:  true,
		}
		events = append(events, e)
	}

	if newtC.OrgID != "" && newtC.OrgID != oldTransConfig.OrgID {
		e := &Event{
			EventSource: fname,
			Key:         "m2mclient.orgid",
			Value:       newtC.OrgID,
			HasUpdated:  true,
		}
		events = append(events, e)
	}

	return events
}
