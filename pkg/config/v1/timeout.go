package v1

import "time"

// 未显式配置时使用的默认超时（毫秒）
const (
	DefaultHTTPTimeoutMS = 30000
	DefaultDBTimeoutMS   = 60000
)

// TimeoutConfig 本地网关侧的超时兜底配置，单位统一为毫秒。
// 平台下发的 timeout 优先级更高，这里只负责"平台没给值"和"值超出本地允许范围"两种情况。
type TimeoutConfig struct {
	// HTTPDefaultMS 平台未下发 timeout 时使用的超时，避免出现 0（永不超时）
	HTTPDefaultMS int
	// HTTPMaxMS 非 0 时对外部传入的超时做上限截断
	HTTPMaxMS int
	// DBMS 数据库连接与查询的整体超时
	DBMS int
}

func (c *TimeoutConfig) Complete() {
	if c.HTTPDefaultMS <= 0 {
		c.HTTPDefaultMS = DefaultHTTPTimeoutMS
	}
	if c.HTTPMaxMS < 0 {
		c.HTTPMaxMS = 0
	}
	if c.DBMS <= 0 {
		c.DBMS = DefaultDBTimeoutMS
	}
}

// HTTPTimeout 返回单次 HTTP 调用最终生效的超时。
// externalMS 是平台下发的毫秒值，非正数表示平台没有指定，此时使用本地默认值。
func (c *TimeoutConfig) HTTPTimeout(externalMS int) time.Duration {
	ms := externalMS
	if ms <= 0 {
		ms = c.HTTPDefaultMS
	}
	if c.HTTPMaxMS > 0 && ms > c.HTTPMaxMS {
		ms = c.HTTPMaxMS
	}
	if ms <= 0 {
		ms = DefaultHTTPTimeoutMS
	}
	return time.Duration(ms) * time.Millisecond
}

// DBTimeout 返回单次数据库操作（连接 + 查询）的超时。
func (c *TimeoutConfig) DBTimeout() time.Duration {
	ms := c.DBMS
	if ms <= 0 {
		ms = DefaultDBTimeoutMS
	}
	return time.Duration(ms) * time.Millisecond
}
