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

func TestUMRequest(t *testing.T) {
	body := ""
	var gereqbody interface{}
	err := json.Unmarshal([]byte(body), &gereqbody)
	if err != nil {
		t.Logf("error: %v", err)
	}
	body = `{"t1": {"a": "b"}, "t2": {"c": "d"}}`
	var gereqbody2 testStruct
	err = json.Unmarshal([]byte(body), &gereqbody2)
	if err != nil {
		require.NoError(t, err)
	}
	body = `{"t1": "", "t2": {"c": "d"}}`
	err = json.Unmarshal([]byte(body), &gereqbody2)
	if err != nil {
		require.NoError(t, err)
	}

	t.Logf("%+v", gereqbody2)

}
