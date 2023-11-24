package main

import (
	"context"
	"encoding/json"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
	"go.uber.org/zap"
)

type IPaaSAgentProtocol struct {
	Headers Headers `json:"headers"`
	Body    Body    `json:"body"`
}

type Headers struct {
	SpecVersion     string `json:"specVersion"`
	ConnectorCorpId string `json:"connectorCorpId"`
	Type            string `json:"type"`
	// connector property
	ConnectorId string `json:"connectorId"`
	ActionId    string `json:"actionId"`
}

type Body struct {
	HTTPRequest  HTTPRequest       `json:"httpRequest"`
	ConfigParams map[string]string `json:"configParams"`
	ConfigId     string            `json:"configId"`
}

type HTTPRequest struct {
	Headers     map[string]string `json:"headers"`
	Method      string            `json:"method"`
	Body        string            `json:"body"`
	ContentType string            `json:"contentType"`
	URL         string            `json:"url"`
	Timeout     int               `json:"timeout"`
}

func (ap IPaaSAgentProtocol) MarshalJSON() ([]byte, error) {
	type Alias IPaaSAgentProtocol
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&ap),
	})
}

func WrapStreamResponseWithString(data string) *payload.DataFrameResponse {
	logger := zap.L()
	logger.Info("wrap stream response with string", zap.String("data", data))
	response, err := json.Marshal(map[string]string{"response": data})
	if err != nil {
		logger.Error("marshal response error", zap.Error(err))
		return nil
	}
	return &payload.DataFrameResponse{
		Code: 200,
		Headers: payload.DataFrameHeader{
			"ContentType": "application/json",
		},
		Message: "success",
		Data:    string(response),
	}
}

func WrapStreamResponseWithBytes(data []byte) *payload.DataFrameResponse {
	return WrapStreamResponseWithString(string(data))
}

func HandleIpaasCallBack(c context.Context, df *payload.DataFrame) (*payload.DataFrameResponse, error) {
	logger := zap.L()
	var data interface{}
	err := json.Unmarshal([]byte(df.Data), &data)
	if err != nil {
		logger.Error("unmarshal data error", zap.Error(err))
		return nil, err
	}
	dataByte, err := json.Marshal(data)
	if err != nil {
		logger.Error("marshal data error", zap.Error(err))
		return nil, err
	}
	ap := &IPaaSAgentProtocol{}
	err = json.Unmarshal(dataByte, &ap)
	if err != nil {
		logger.Error("unmarshal agent protocol error: ", zap.Error(err))
		return nil, err
	}
	logger.Info("IPaaSAgentProtocol: ",
		zap.Any("Headers: ", ap.Headers),
		zap.Any("Body: ", ap.Body),
	)

	switch ap.Headers.Type {
	case "HTTP":
		resp, err := HandleHTTPRequest(ap.Body.HTTPRequest)
		if err != nil {
			logger.Error("handle http request error: ", zap.Error(err))
			return nil, err
		}
		return WrapStreamResponseWithBytes(resp), nil
	case "MYSQL":
		resp, err := HandleMySQLProxyRequest(ap)
		if err != nil {
			logger.Error("handle mysql request error: ", zap.Error(err))
			return nil, err
		}
		return WrapStreamResponseWithBytes(resp), nil

	// case "config":
	// 	return HandleConfig(agentProtocol)
	default:
		logger.Error("unknown type: ", zap.String("type", ap.Headers.Type))
		return nil, nil
	}
}
