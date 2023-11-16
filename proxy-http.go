package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

var httpClient = &http.Client{}

func parseHTTPAgentResponse(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	logger := zap.S()
	logger.Info("http request success", zap.String("response", string(body)))
	m := map[string]interface{}{
		"status":     resp.Status,
		"statusCode": resp.StatusCode,
		"proto":      resp.Proto,
		"header":     resp.Header,
		"body":       string(body),
	}

	// 将m编码为JSON，并防止转义
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(m)

	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func HandleHTTPRequest(ipaasHTTPRequest HTTPRequest) ([]byte, error) {
	logger := zap.S()
	var gwReqBody interface{}
	err := json.Unmarshal([]byte(ipaasHTTPRequest.Body), &gwReqBody)

	if err != nil {
		logger.Error("unmarshal request body error", zap.Error(err))
		return nil, err
	}
	// gwReqBodyMap, ok := gwReqBody.(map[string]interface{})
	// if !ok {
	// 	return nil, fmt.Errorf("[HandleHTTPRequest] invalid request, %v", gwReqBody)
	// }

	method := ipaasHTTPRequest.Method
	url := ipaasHTTPRequest.URL

	// bodyStr, ok := gwReqBodyMap["Body"].(string)
	// if !ok {
	// 	return nil, fmt.Errorf("[HandleHTTPRequest] invalid request body")
	// }

	// body := strings.NewReader(bodyStr)
	body := strings.NewReader(ipaasHTTPRequest.Body)
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		logger.Error("create http request error", zap.Error(err))
		return nil, err
	}

	// for key, value := range agentProtocol.Body.Header {
	// 	request.Header.Set(key, value)
	// }
	// query := request.URL.Query()
	// for key, value := range agentProtocol.Body.Query {
	// 	query.Set(key, value)
	// }
	// request.URL.RawQuery = query.Encode()
	// if agentProtocol.Body.Method == "POST" || agentProtocol.Body.Method == "PUT" {
	// 	request.Header.Set("Content-Type", "application/json")
	// }
	ctx, cancel := context.WithTimeout(request.Context(), 5*time.Second)
	defer cancel()
	request = request.WithContext(ctx)
	response, err := httpClient.Do(request)
	if err != nil {
		logger.Error("http request error", zap.Error(err))
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}
	m, err2 := parseHTTPAgentResponse(response)
	return m, err2
}
