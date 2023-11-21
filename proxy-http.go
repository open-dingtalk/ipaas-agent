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
	logger := zap.L()
	logger.Info("http request success: ", zap.String("response", string(body)))
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
	logger := zap.L()
	var gwReqBody interface{}
	err := json.Unmarshal([]byte(ipaasHTTPRequest.Body), &gwReqBody)

	if err != nil {
		logger.Error("unmarshal request body error: ", zap.Error(err))
		return nil, err
	}

	method := ipaasHTTPRequest.Method
	url := ipaasHTTPRequest.URL

	body := strings.NewReader(ipaasHTTPRequest.Body)
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		logger.Error("create http request error: ", zap.Error(err))
		return nil, err
	}

	for key, value := range ipaasHTTPRequest.Headers {
		request.Header.Set(key, value)
	}
	request.Header.Set("Content-Type", "application/json")
	if ipaasHTTPRequest.ContentType != "" {
		request.Header.Set("Content-Type", ipaasHTTPRequest.ContentType)
	}
	timeout := ipaasHTTPRequest.Timeout
	if timeout == 0 {
		timeout = 5
	}

	ctx, cancel := context.WithTimeout(request.Context(), time.Duration(timeout)*time.Second)
	defer cancel()
	request = request.WithContext(ctx)
	response, err := httpClient.Do(request)
	if err != nil {
		logger.Error("http request error: ", zap.Error(err))
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}
	m, err2 := parseHTTPAgentResponse(response)
	return m, err2
}
