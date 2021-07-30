package core

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"k8s.io/klog/v2"

	"github.com/edgehook/ctrlapp/pkg/common/config"
	"github.com/edgehook/ctrlapp/pkg/core/mqtt"
	"github.com/edgehook/ctrlapp/pkg/utils"
)

type Transport struct {
	deviceID      string
	conf          *config.AControlConfig
	client        *mqtt.Client
	requestsQueue chan *Request
}

var ()

func NewTransport() *Transport {
	macs := utils.GetLocalMACs()
	if macs == nil || len(macs) == 0 {
		//maybe, we will replace it as a UUID.
		klog.Errorf("no mac address, please check your device.")
		return nil
	}

	appCtrlConf := config.GetAControlConfig()
	if appCtrlConf == nil {
		klog.Errorf("invalid conf, please check your config file.")
		return nil
	}

	t := &Transport{
		requestsQueue: make(chan *Request, 128),
	}

	t.deviceID = macs[0]
	t.conf = appCtrlConf

	if t.deviceID == "" {
		t.deviceID = appCtrlConf.MqttConfig.ClientID
	}

	return t
}

//StartUp.
func (t *Transport) Run() error {
	//will message
	wm := &mqtt.WillMessage{
		Topic:    t.buildWillTopic(),
		Payload:  t.buildWillMessage(),
		Qos:      byte(t.conf.MqttConfig.QOS),
		Retained: true,
	}

	//tls config
	tlsConfig, err := CreateTLSConfig(t.conf.MqttConfig.CaFilePath,
		t.conf.MqttConfig.CertFilePath, t.conf.MqttConfig.KeyFilePath)
	if err != nil {
		tlsConfig = nil
	}

	//create the mqtt client.
	brokerAddress := t.conf.MqttConfig.Broker
	if !strings.Contains(brokerAddress, ":") {
		brokerAddress = brokerAddress + ":1883"
	}

	usr := t.conf.MqttConfig.User
	pwd := t.conf.MqttConfig.Passwd
	t.client = mqtt.NewMQTTClient(brokerAddress, t.deviceID, usr, pwd, tlsConfig, wm)
	t.client.OnConnectFn = t.OnConnect
	t.client.OnLostFn = t.OnLost

	//if connect failed, we always retry to connect.
retry_connect:
	err = t.client.Connect()
	if err != nil {
		time.Sleep(3 * time.Second)
		goto retry_connect
	}

	klog.Infof("connect to broker %s successfuly", brokerAddress)
	return nil
}

//buildWillTopic build will topic.
func (t *Transport) buildWillTopic() string {
	topic := "device/will/" +
		t.deviceID + "/delete/willmessage"

	return topic
}

func (t *Transport) buildWillMessage() string {
	message := "{}"

	return message
}

func (t *Transport) GetRequestFromQueue() *Request {
	req, ok := <-t.requestsQueue
	if !ok {
		return nil
	}

	return req
}

// create tls config
func CreateTLSConfig(caFile, certFile, keyFile string) (*tls.Config, error) {
	pool := x509.NewCertPool()
	rootCA, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	ok := pool.AppendCertsFromPEM(rootCA)
	if !ok {
		return nil, fmt.Errorf("fail to load ca content")
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		ClientCAs:    pool,
		ClientAuth:   tls.RequestClientCert,
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
	}

	return tlsConfig, nil
}

//Send data over mqtt bus.
func (t *Transport) Send(topic, payload string) error {
	return t.client.Publish(topic, payload, byte(t.conf.MqttConfig.QOS), false)
}

func (t *Transport) OnConnect(c *mqtt.Client) {
	klog.Infof("On Connected!")

	//subscribe mqtt topic.
	subTopic := "edge/+/" + t.deviceID + "/#"
	err := t.client.Subscribe(subTopic, byte(t.conf.MqttConfig.QOS), t.MessageArrived)
	if err != nil {
		klog.Infof("subscribe with err %s", err.Error())
	}

	klog.Infof("subscribe topic: [%s]", subTopic)
}

func (t *Transport) OnLost(c *mqtt.Client, err error) {
	klog.Infof("Connect is lost with ther err = [%s]!", err.Error())

	/*
	*  notify the m2mcore into offline state.
	*  TODO:
	 */
}

func (t *Transport) MessageArrived(topic string, payload []byte) {

	//klog.Infof("MessageArrived topic= %s, payload = %s", topic, string(payload))

	matchControl := "edge/control/" + t.deviceID
	matchResp := "edge/response/"
	if strings.HasPrefix(topic, matchControl) {
		//this is a comtrol request.
		request := ParseRequest(topic, string(payload))
		if request != nil {
			t.requestsQueue <- request
		}
	} else if strings.HasPrefix(topic, matchResp) {
		resp := ParseResponse(topic, string(payload))
		if resp != nil {

		}
	}
}
