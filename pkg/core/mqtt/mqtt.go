package mqtt

import (
	"crypto/tls"
	"errors"
	_ "log"
	_ "os"
	"time"

	"k8s.io/klog/v2"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgehook/ctrlapp/pkg/common/crypto"
)

/*
* Will message.
 */
type WillMessage struct {
	Topic    string
	Payload  string
	Qos      byte
	Retained bool
}

/*
* Mqtt Client
 */
type Client struct {
	// scheme://host:port
	// Where "scheme" is one of "tcp", "ssl", or "ws", "host" is the ip-address (or hostname)
	// and "port" is the port on which the broker is accepting connections.
	Host         string
	User, Passwd string
	// the client id to be used by this client when
	// connecting to the MQTT broker.
	ClientID string
	//0: QOSAtMostOnce, 1: QOSAtLeastOnce, 2: QOSExactlyOnce.
	QOS byte
	//if the flag set true, server will store the message and
	//can be delivered to future subscribers.
	Retain bool
	//the state of client.
	State string
	// tls config
	tlsConfig *tls.Config
	client    paho.Client

	OnConnectFn func(*Client)
	OnLostFn    func(*Client, error)
}

func NewMQTTClient(host, clientID, usr, pwd string,
	tlsConfig *tls.Config, wm *WillMessage) *Client {

	// mqtt client.
	client := &Client{
		Host:        host,
		ClientID:    clientID,
		QOS:         0,
		Retain:      false,
		tlsConfig:   tlsConfig,
		OnConnectFn: nil,
		OnLostFn:    nil,
	}

	options := paho.NewClientOptions()

	options.AddBroker(host)
	options.SetClientID(clientID)

	options.SetUsername(usr)
	//decrypt the passwd.
	passwd, _ := crypto.Decrypt(pwd)
	options.SetPassword(string(passwd))

	//clean session is true by default.
	// use memory store by default.

	if tlsConfig != nil {
		options.SetTLSConfig(tlsConfig)
	}

	options.SetOnConnectHandler(client.OnConnect)
	options.SetConnectionLostHandler(client.OnLost)

	/*
	* guarantee order of the incoming message.
	 */
	options.SetOrderMatters(true)

	//how long tome to send ping request.
	options.SetKeepAlive(15 * time.Second)
	//how long time to recieve the pingresp after send ping request.
	options.SetPingTimeout(10 * time.Second)
	options.SetMessageChannelDepth(125)

	/*
	* Setting Will topic
	 */
	if wm != nil {
		cryptedMsg, _ := crypto.Encrypt([]byte(wm.Payload))
		options.SetWill(wm.Topic, cryptedMsg, wm.Qos, wm.Retained)
	}
	options.SetAutoReconnect(true)

	c := paho.NewClient(options)
	client.client = c

	//add debug logger for paho mqtt
	/*paho.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	paho.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	paho.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	paho.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)*/

	return client
}

/*
* Connect the mqtt broker.
 */
func (c *Client) Connect() error {
	if c.client == nil {
		return errors.New("nil client")
	}

	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

/*
* Publish the message.
 */
func (c *Client) Publish(topic, payload string, qos byte, retained bool) error {
	// klog.Infof("topic: %s, payload: %s", topic, payload)
	if topic == "" {
		return errors.New("invaliad topic")
	}

	if c.client == nil {
		return errors.New("nil client")
	}

	//Is connected ?
	if !c.client.IsConnectionOpen() {
		return errors.New("connection is not active")
	}

	//encrypt the message.
	crypted, err := crypto.Encrypt([]byte(payload))
	if err != nil {
		return err
	}

	//publish the message
	if token := c.client.Publish(topic, qos, retained, crypted); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

/*
* Subscribe the message.
 */
func (c *Client) Subscribe(topic string, qos byte, fn func(tpc string, payload []byte)) error {

	if c.client == nil {
		return errors.New("nil client")
	}

	if !c.client.IsConnectionOpen() {
		return errors.New("connection is not active")
	}

	callback := func(client paho.Client, message paho.Message) {
		rawData := message.Payload()

		//decrypt the message.
		payload, err := crypto.Decrypt(string(rawData))
		if err != nil {
			klog.Infof("error message format, Ignored!")
			return
		}

		fn(message.Topic(), payload)
	}

	//subscribe the topic.
	if token := c.client.Subscribe(topic, qos, callback); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (c *Client) IsConnected() bool {
	return c.client.IsConnected()
}

/*
* Client callback when on connect.
 */
func (c *Client) OnConnect(client paho.Client) {
	if c.OnConnectFn != nil {
		c.OnConnectFn(c)
	}
}

/*
* Client callback when on lost.
 */
func (c *Client) OnLost(client paho.Client, err error) {
	// issue:
	//EOF / read: connection reset by peer
	if c.OnLostFn != nil {
		c.OnLostFn(c, err)
	}
}

/*
* Unsubscribe
 */
func (c *Client) Unsubscribe(topics string) error {
	if c.client == nil {
		return errors.New("nil client")
	}

	if !c.client.IsConnectionOpen() {
		return errors.New("connection is not active")
	}

	if token := c.client.Unsubscribe(topics); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

/*
* Close.
 */
func (c *Client) Close() {
	c.client.Disconnect(250)
}
