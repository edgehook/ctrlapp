package core

import (
	"encoding/json"
	"github.com/satori/go.uuid"
	"strings"
)

type Request struct {
	Cmd     string
	MsgID   string
	Content string
}

func BuildRequest(cmd string, content interface{}) *Request {
	var data string

	switch content.(type) {
	case []byte:
		data = string(content.([]byte))
	case string:
		data = content.(string)
	default:
		jsonData, _ := json.Marshal(content)
		data = string(jsonData)
	}

	r := &Request{
		Cmd:     cmd,
		MsgID:   uuid.NewV4().String(),
		Content: data,
	}

	return r
}

func (r *Request) GetMsgID() string {
	return r.MsgID
}

func (r *Request) GetMsgContent() string {
	return r.Content
}

func (r *Request) BuildTopic(macAddr string) string {

	topic := "device/report/" + macAddr + "/" + r.Cmd + "/" + r.MsgID
	return topic
}

func (r *Request) BuildPayload() string {
	return r.Content
}

/*
* Parse the request package.
 */
func ParseRequest(topic, payload string) *Request {
	//parse topic
	levels := strings.Split(topic, "/")
	if len(levels) < 5 {
		return nil
	}

	r := &Request{
		Cmd:     levels[3],
		MsgID:   levels[4],
		Content: payload,
	}

	return r
}

/*
* Build Response message.
 */
func (r *Request) BuildResponse(code int, reason string, parms interface{}) *Response {
	result := &RespResult{
		Code:   code,
		Reason: reason,
		Parms:  parms,
	}
	payload, _ := json.Marshal(result)

	response := &Response{
		ParentID: r.MsgID,
		Content:  string(payload),
	}

	return response
}

type Response struct {
	ParentID string
	//message body
	Content string
}

type RespResult struct {
	Code   int         `json:"errorcode"`
	Reason string      `json:"reason"`
	Parms  interface{} `json:"parameter"`
}

/*
* ParseResponse
* parse the async response message.
 */
func ParseResponse(topic, payload string) *Response {
	//parse topic
	levels := strings.Split(topic, "/")
	if len(levels) < 6 {
		return nil
	}

	r := &Response{
		ParentID: levels[3],
		Content:  payload,
	}

	return r
}

func (r *Response) BuildTopic(macAddr string) string {
	topic := "device/response/" + macAddr + "/" + r.ParentID
	return topic
}

func (r *Response) BuildPayload() string {
	return r.Content
}
