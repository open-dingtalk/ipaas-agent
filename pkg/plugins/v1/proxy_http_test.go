package v1_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/open-dingtalk/ipaas-agent/pkg/plugins/v1"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// slowServer 返回一个固定延迟的 HTTP 服务，用来真实触发超时分支。
func slowServer(delay time.Duration) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
}

// TestHandleHTTPRequestUsesConfiguredTimeout 覆盖上游 issue #1：
// 平台没有下发 timeout 时，超时由本地配置决定，而不是一句写死的 5 秒。
func TestHandleHTTPRequestUsesConfiguredTimeout(t *testing.T) {
	t.Cleanup(func() {
		viper.Set("timeout.http_default_ms", 0)
		viper.Set("timeout.http_max_ms", 0)
	})

	// 后端比本地默认超时慢：必须失败，而不是无限等下去
	slow := slowServer(400 * time.Millisecond)
	defer slow.Close()
	viper.Set("timeout.http_default_ms", 100)
	viper.Set("timeout.http_max_ms", 0)

	start := time.Now()
	resp, err := v1.HandleHTTPRequest(v1.HTTPRequest{Method: http.MethodGet, URL: slow.URL})
	require.Nil(t, resp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "context deadline exceeded")
	require.Less(t, time.Since(start), 2*time.Second, "应在配置的超时处结束，而不是等后端返回")

	// 放宽本地默认超时后，同一个后端应当正常返回
	viper.Set("timeout.http_default_ms", 3000)
	resp, err = v1.HandleHTTPRequest(v1.HTTPRequest{Method: http.MethodGet, URL: slow.URL})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, `{"ok":true}`, resp.Body)
}

// TestHandleHTTPRequestHonoursPlatformTimeout 说明平台下发的值仍然优先，
// 本地配置只做兜底和上限截断。
//
// 注意单位：v1 协议的 timeout 字段是"秒"（proxy_http.go 按秒换算），
// 只有 v2 的 http_plugin 按毫秒处理，两个协议共用同一个字段名。
func TestHandleHTTPRequestHonoursPlatformTimeout(t *testing.T) {
	t.Cleanup(func() {
		viper.Set("timeout.http_default_ms", 0)
		viper.Set("timeout.http_max_ms", 0)
	})
	viper.Set("timeout.http_default_ms", 60000)

	slow := slowServer(1500 * time.Millisecond)
	defer slow.Close()

	// 平台给的是 1 秒，后端要 1.5 秒：必须按平台的 1 秒失败，本地 60 秒默认值不参与
	start := time.Now()
	_, err := v1.HandleHTTPRequest(v1.HTTPRequest{Method: http.MethodGet, URL: slow.URL, Timeout: 1})
	require.Error(t, err)
	require.Less(t, time.Since(start), 3*time.Second, "应按平台下发的 1 秒结束")

	// 上限截断：平台给了 600 秒，本地最多允许 0.5 秒
	viper.Set("timeout.http_max_ms", 500)
	start = time.Now()
	_, err = v1.HandleHTTPRequest(v1.HTTPRequest{Method: http.MethodGet, URL: slow.URL, Timeout: 600})
	require.Error(t, err)
	require.Less(t, time.Since(start), time.Second, "应按本地上限 0.5 秒截断")
}
