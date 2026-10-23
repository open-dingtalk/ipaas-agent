package v1_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

type testStruct struct {
	T1 map[string]string `json:"t1,omitempty"`
	T2 map[string]string `json:"t2"`
}

// TestUnmarshalRequestBody 固化外部报文反序列化的三条真实约束：
//
//  1. 空字符串不是合法 JSON，DataFrame.Data 为空时必须走错误分支；
//  2. 字段形态正确时按预期解析；
//  3. 平台有时会把本应是对象的字段下发成空字符串（例如 "headers": ""），
//     此时 encoding/json 直接报错而不是忽略它。
//
// 第 3 条正是这条用例存在的原因：所有反序列化外部报文的路径都必须把错误当可恢复错误
// 处理并回给平台，既不能 panic，也不能静默当成空对象。
func TestUnmarshalRequestBody(t *testing.T) {
	var empty interface{}
	require.Error(t, json.Unmarshal([]byte(""), &empty), "空报文不是合法 JSON")

	var valid testStruct
	require.NoError(t, json.Unmarshal([]byte(`{"t1": {"a": "b"}, "t2": {"c": "d"}}`), &valid))
	require.Equal(t, map[string]string{"a": "b"}, valid.T1)
	require.Equal(t, map[string]string{"c": "d"}, valid.T2)

	var lenient testStruct
	err := json.Unmarshal([]byte(`{"t1": "", "t2": {"c": "d"}}`), &lenient)
	require.Error(t, err, "空字符串不能反序列化为 map，必须报错")
	t.Logf("空字符串作为 map 字段时的错误: %v", err)
}
