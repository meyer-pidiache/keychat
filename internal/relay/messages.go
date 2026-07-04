package relay

import "encoding/json"

type OKResponse struct {
	Type    string `json:"type"`
	EventID string `json:"event_id"`
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

type EOSEResponse struct {
	Type  string `json:"type"`
	SubID string `json:"subscription"`
}

type ClosedResponse struct {
	Type    string `json:"type"`
	SubID   string `json:"subscription"`
	Message string `json:"message"`
}

type NoticeResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func NewOK(subID, eventID string, ok bool, msg string) []byte {
	resp := []interface{}{"OK", eventID, ok, msg}
	data, _ := json.Marshal(resp)
	return data
}

func NewEOSE(subID string) []byte {
	resp := []interface{}{"EOSE", subID}
	data, _ := json.Marshal(resp)
	return data
}

func NewClosed(subID, msg string) []byte {
	resp := []interface{}{"CLOSED", subID, msg}
	data, _ := json.Marshal(resp)
	return data
}

func NewNotice(msg string) []byte {
	resp := []interface{}{"NOTICE", msg}
	data, _ := json.Marshal(resp)
	return data
}

func NewEventMessage(subID string, event *Event) []byte {
	resp := []interface{}{"EVENT", subID, event}
	data, _ := json.Marshal(resp)
	return data
}
