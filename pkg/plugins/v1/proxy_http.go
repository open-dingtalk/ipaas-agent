package v1

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/open-dingtalk/ipaas-agent/pkg/logger"
)

type HTTPRequest struct {
	Headers     map[string]string `json:"headers"`
	Method      string            `json:"method"`
	Body        string            `json:"body"`
	ContentType string            `json:"contentType"`
	URL         string            `json:"url"`
	Timeout     int               `json:"timeout"`
}

type HTTPResponse struct {
	Status     string            `json:"status"`
	StatusCode int               `json:"statusCode"`
	Proto      string            `json:"proto"`
	Header     map[string]string `json:"header"`
	Body       string            `json:"body"`
}

var httpClient = &http.Client{}

func parseHTTPAgentResponse(resp *http.Response, respv1 *HTTPResponse) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	logger.Log1.Infof("HTTP 请求成功: %s", string(body))
	respv1.Status = resp.Status
	respv1.StatusCode = resp.StatusCode
	respv1.Proto = resp.Proto
	respv1.Header = make(map[string]string)
	for k, v := range resp.Header {
		respv1.Header[k] = strings.Join(v, ",")
	}
	respv1.Body = string(body)
	return nil
}

func HandleHTTPRequest(ipaasHTTPRequest HTTPRequest) (*HTTPResponse, error) {
	method := ipaasHTTPRequest.Method
	url := ipaasHTTPRequest.URL

	body := strings.NewReader(ipaasHTTPRequest.Body)
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		logger.Log1.Errorf("create http request error: %v", err)
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
		logger.Log1.Errorf("http request error: %v", err)
		return nil, err
	}
	defer response.Body.Close()
	var m HTTPResponse // HTTPResponse
	err = parseHTTPAgentResponse(response, &m)
	if err != nil {
		logger.Log1.Errorf("parse http response error: %v", err)
		return nil, err
	}
	return &m, err
}
