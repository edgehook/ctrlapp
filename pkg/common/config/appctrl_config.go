package config

import (
	"github.com/edgehook/ctrlapp/pkg/utils"
	"os"
	"sync"
)

const (
	EnvironmentalConfigPath = "QCTRL_CONFIG_PATH"
)

var (
	once         = sync.Once{}
	ACTRL_CONFIG *Config
)

type AControlAppConfig struct {
	OrgID       string
	ServerEPID  string
	ServerAppID string
	ClientAppID string
}

type AControlMqttConfig struct {
	Broker       string
	User, Passwd string
	ClientID     string
	DeviceUUID   string
	DeviceName   string
	// QOS indicates mqtt qos
	// 0: QOSAtMostOnce, 1: QOSAtLeastOnce, 2: QOSExactlyOnce
	// default 0
	// Note: Can not use "omitempty" option,  It will affect the output of the default configuration file
	QOS          int
	CaFilePath   string
	CertFilePath string
	KeyFilePath  string
	IsSwitched   bool
}

type AControlConfig struct {
	InstallPath string
	AppConfig   AControlAppConfig
	MqttConfig  AControlMqttConfig
	isLoad      bool
}

var aConfig = &AControlConfig{}

func GetAControlConfig() *AControlConfig {

	if aConfig.isLoad {
		return aConfig
	}

	aConfig.InstallPath = ACTRL_CONFIG.GetString("ctrlapp.install-path")

	//App config
	aConfig.AppConfig.OrgID = ACTRL_CONFIG.GetString("ctrlapp.app.orgid")
	aConfig.AppConfig.ServerEPID = ACTRL_CONFIG.GetString("ctrlapp.app.server-endpoint-id")
	aConfig.AppConfig.ServerAppID = ACTRL_CONFIG.GetString("ctrlapp.app.server-application-id")
	aConfig.AppConfig.ClientAppID = ACTRL_CONFIG.GetString("ctrlapp.app.client-application-id")

	//Mqtt config
	aConfig.MqttConfig.QOS = ACTRL_CONFIG.GetInt("ctrlapp.mqtt.qos")
	aConfig.MqttConfig.Broker = ACTRL_CONFIG.GetString("ctrlapp.mqtt.broker")
	aConfig.MqttConfig.User = ACTRL_CONFIG.GetString("ctrlapp.mqtt.usr")
	aConfig.MqttConfig.Passwd = ACTRL_CONFIG.GetString("ctrlapp.mqtt.passwd")
	aConfig.MqttConfig.ClientID = ACTRL_CONFIG.GetString("ctrlapp.mqtt.clientid")
	aConfig.MqttConfig.DeviceUUID = ACTRL_CONFIG.GetString("ctrlapp.mqtt.deviceuuid")
	aConfig.MqttConfig.DeviceName = ACTRL_CONFIG.GetString("ctrlapp.mqtt.device-name")
	aConfig.MqttConfig.CaFilePath = ACTRL_CONFIG.GetString("ctrlapp.mqtt.cafile")
	aConfig.MqttConfig.CertFilePath = ACTRL_CONFIG.GetString("ctrlapp.mqtt.certfile")
	aConfig.MqttConfig.KeyFilePath = ACTRL_CONFIG.GetString("ctrlapp.mqtt.keyfile")
	switched := ACTRL_CONFIG.GetString("ctrlapp.mqtt.switched")
	if switched == "yes" {
		aConfig.MqttConfig.IsSwitched = true
	} else {
		aConfig.MqttConfig.IsSwitched = false
	}

	aConfig.isLoad = true

	return aConfig
}

func SaveAControlSwitchedConfig(val bool) error {
	defer func() {
		//update the M2mConfig.
		ACTRL_CONFIG.ReloadConfig()
		aConfig.isLoad = false
	}()

	if val {
		ACTRL_CONFIG.SetString("ctrlapp.mqtt.switched", "yes")
	} else {
		ACTRL_CONFIG.SetString("ctrlapp.mqtt.switched", "no")
	}

	//save config to conf/lwm2m.yaml
	return ACTRL_CONFIG.SaveConfig()
}

func SaveAControlDeviceName(val string) error {
	ACTRL_CONFIG.SetString("ctrlapp.mqtt.device-name", val)
	//save config to conf/lwm2m.yaml
	return ACTRL_CONFIG.SaveConfig()
}

func getConfigDirectory() string {
	//get env
	config := os.Getenv(EnvironmentalConfigPath)
	if config != "" {
		return config
	}

	return utils.GetInstallRootPath() + "/conf"
}

func LoadConfig() {
	once.Do(func() {
		//load the config.yaml from conf/
		configDir := getConfigDirectory()
		ACTRL_CONFIG = NewYamlConfig(configDir, "config.yaml")
		aConfig.isLoad = false
	})
}
