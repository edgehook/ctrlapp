package core

import (
	"encoding/json"
	"github.com/edgehook/ctrlapp/pkg/common/config"
	"github.com/edgehook/ctrlapp/pkg/utils"
	"k8s.io/klog/v2"
	"path/filepath"
	"strings"
)

type CtrAppCore struct {
	isSwitched   bool
	isNeedReport bool
	conf         *config.AControlConfig
	transport    *Transport
}

var (
	core = &CtrAppCore{}
)

func DoStartUpCore() {
	config.LoadConfig()

	core.conf = config.GetAControlConfig()
	t := NewTransport()
	if t == nil {
		return
	}

	// connect the broker.
	core.transport = t
	err := core.transport.Run()
	if err != nil {
		klog.Fatalf("core.transport.Run with err :%v", err)
	}

	//Configure the lwm2m.
	InstallPath := core.GetLwm2mInstallPath()
	lwm2mConfigyaml := filepath.Join(InstallPath, "conf", "edge.yaml")

	config.LoadM2mConfig(lwm2mConfigyaml)

	if !core.isSwitched {
		/*
		* If no switch action occured, we always use the mqtt config
		* from appctrl's config yaml. Or. use the lwm2m own config yaml.
		 */
		lwm2mConfig := &config.TransportConfig{
			Broker:       core.conf.MqttConfig.Broker,
			User:         core.conf.MqttConfig.User,
			Passwd:       core.conf.MqttConfig.Passwd,
			ClientID:     core.transport.deviceID,
			DeviceName:   core.conf.MqttConfig.DeviceName,
			QOS:          core.conf.MqttConfig.QOS,
			CaFilePath:   core.conf.MqttConfig.CaFilePath,
			CertFilePath: core.conf.MqttConfig.CertFilePath,
			KeyFilePath:  core.conf.MqttConfig.KeyFilePath,
			ServerEPID:   core.conf.AppConfig.ServerEPID,
			ServerAppID:  core.conf.AppConfig.ServerAppID,
			ClientAppID:  core.conf.AppConfig.ClientAppID,
			OrgID:        core.conf.AppConfig.OrgID,
		}

		config.SaveTransportConfig(lwm2mConfig)
	}

	//Start lwm2m.
	core.StartLwm2m()

	//Monitor the lwm2m yaml change.
	config.AddWatchConfig(core.OnYamlChanged)

	//Loop recieve the request.
	for {
		req := core.transport.GetRequestFromQueue()
		if req == nil {
			continue
		}

		switch req.Cmd {
		case "config":
			klog.Infof("config and restart the lwm2m.")
			core.DoConfigLwm2m(req)
		case "changename":
			klog.Infof("Change name")
			core.DoChangeName(req)
		case "getuuid":
			klog.Infof("Get uuid")
			core.DoGetDeviceUUID(req)
		}
	}
}

func (c *CtrAppCore) DoConfigLwm2m(req *Request) {
	configString := req.GetMsgContent()

	oldConfig := config.GetTransportConfig()
	c.isNeedReport = false
	err := config.SetTransportConfig(configString)
	if err != nil {
		klog.Errorf("SetTransportConfig with err %v", err)
		c.SendResponse(req, 1, "saved config file failed", "")
		c.isNeedReport = true
		return
	}
	newConfig := config.GetTransportConfig()

	//switche lwm2m to anthor server
	if oldConfig.IsNeedSwitch(newConfig) {

		//Start lwm2m.
		c.StartLwm2m()

		//Mark this switch flag.
		c.isSwitched = true
		config.SaveAControlSwitchedConfig(true)
	}

	c.SendResponse(req, 0, "success", "")
}

func (c *CtrAppCore) DoChangeName(req *Request) {
	jsonString := req.GetMsgContent()

	parms := make(map[string]interface{})

	err := json.Unmarshal([]byte(jsonString), &parms)
	if err != nil {
		klog.Infof("Unmarshal with err %v", err)
		c.SendResponse(req, 1, "Unmarshal failed", err.Error())
		return
	}

	v, ok := parms["devicename"]
	if !ok {
		klog.Infof("invalid request")
		c.SendResponse(req, 1, "invalid request", "")
		return
	}

	deviceName, _ := v.(string)

	/**
	* if devicename is empty, we don't set it and return
	* error
	 */
	if deviceName == "" {
		klog.Errorf("device name must bo not empty.")
		c.SendResponse(req, 2, "device name must bo not empty.", "")
		return
	}

	c.isNeedReport = false
	err = config.SetDeviceName(deviceName)
	if err != nil {
		c.isNeedReport = true
		c.SendResponse(req, 1, "SetDeviceName failed", err.Error())
		klog.Errorf("SetDeviceName failed with err %v", err)
		return
	}

	//Start lwm2m.
	c.StartLwm2m()

	c.SendResponse(req, 0, "success", "")
}

func (c *CtrAppCore) DoGetDeviceUUID(req *Request) {

	deviceUUID := c.conf.MqttConfig.DeviceUUID

	parms := make(map[string]interface{})
	parms["uuid"] = deviceUUID

	c.SendResponse(req, 0, "success", parms)
}

// SendResponse send response.
func (c *CtrAppCore) SendResponse(req *Request, code int, reason string, parms interface{}) error {
	times := int(0)
	deviceID := c.transport.deviceID
	resp := req.BuildResponse(1, "json Marshal failed.", "")

	bytesData, err := json.Marshal(parms)
	if err != nil {
		klog.Infof("Marshal(parms) with err %v", err)
		goto trans_send
	}

	resp = req.BuildResponse(code, reason, string(bytesData))

trans_send:
	err = c.transport.Send(resp.BuildTopic(deviceID), resp.BuildPayload())
	if err != nil {
		if times > 3 {
			return err
		}

		times++
		goto trans_send
	}

	return nil
}

func (c *CtrAppCore) SendReport(cmd string, content interface{}) error {
	times := int(0)
	deviceID := c.transport.deviceID
	req := BuildRequest(cmd, content)

trans_send:
	err := c.transport.Send(req.BuildTopic(deviceID), req.BuildPayload())
	if err != nil {
		if times > 3 {
			return err
		}

		times++
		goto trans_send
	}

	return nil
}

func (c *CtrAppCore) GetLwm2mInstallPath() string {
	if c.conf.InstallPath != "" {
		return c.conf.InstallPath
	}

	// we are always assume that ctrlappp installed in
	// a same path with lwm2m.
	return utils.GetInstallRootPath()
}

func (c *CtrAppCore) StartLwm2m() error {
	if utils.FileIsExist("/bin/systemctl") {
		output, err := utils.Execute1("systemctl", "restart", "AppHub-Edge")
		if err != nil {
			klog.Errorf("Error message: %s", output)
			return err
		}
	} else if utils.FileIsExist("/usr/sbin/service") {
		output, err := utils.Execute1("service", "restart", "AppHub-Edge")
		if err != nil {
			klog.Errorf("Error message: %s", output)
			return err
		}
	}

	return nil
}

func (c *CtrAppCore) OnYamlChanged(ev *config.Event) {
	if c.isNeedReport == false {
		c.isNeedReport = true
		return
	}

	if strings.Contains(ev.Key, "m2mclient.device-name") {
		klog.Infof("Device Name changed to  %v", ev.Value)
		//Report the Name change event.
		parms := make(map[string]interface{})
		parms["devicename"] = ev.Value
		c.SendReport("namechanged", parms)

		//save the value into config yaml.
		deviceName, _ := ev.Value.(string)
		config.SaveAControlDeviceName(deviceName)
	}
}
