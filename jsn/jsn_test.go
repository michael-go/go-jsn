package jsn

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalToJson(t *testing.T) {
	var j Json
	err := json.Unmarshal([]byte(`{
		"koko": 1, 
		"float": 2.07,
		"moko": "cool",
		"deep": {
			"lala": true,
			"dada": null
		},
		"arr" : [1, "x", null],
		"sarr": ["one", "two"],
		"jarr": [
			{"a": 1},
			{"b": 2}
			]
	}`), &j)
	require.NoError(t, err)

	assert.True(t, j.Exists("koko"))
	assert.False(t, j.Exists("koko2"))

	assert.Equal(t, j.V("moko").String(), String{"cool", true})
	assert.Equal(t, j.V("no").String(), String{"", false})

	assert.Equal(t, j.V("koko").Int(), Int{1, true})
	assert.Equal(t, j.V("no").Int(), Int{0, false})
	assert.Equal(t, j.V("float").Int(), Int{2, true})
	assert.Equal(t, j.V("float").Float64(), Float64{2.07, true})

	assert.Equal(t, j.V("deep").V("lala").Undefined(), false)
	assert.Equal(t, j.V("deep").V("lala").Null(), false)
	assert.Equal(t, j.V("deep").V("lala").Bool(), Bool{true, true})
	assert.Equal(t, j.V("deep").V("dada").Undefined(), false)
	assert.Equal(t, j.V("deep").V("dada").Null(), true)
	assert.Equal(t, j.V("deep").V("no").Undefined(), true)
	assert.Equal(t, j.V("deep").V("no").Null(), false)
	assert.Equal(t, j.V("deep").V("no").NullOrUndefined(), true)

	require.Len(t, j.V("arr").Array().Values(), 3)
	assert.Equal(t, j.V("arr").Array().Values()[0].Int(), Int{1, true})
	assert.Equal(t, j.V("arr").Array().Values()[1].String(), String{"x", true})
	assert.True(t, j.V("arr").Array().Values()[2].Null())

	assert.Equal(t, j.V("arr").I(1).String(), String{"x", true})
	assert.Equal(t, j.V("arr").I(10).String(), String{"", false})

	assert.True(t, j.V("jarr").Array().IsValid)
	assert.Len(t, j.V("jarr").Array().Values(), 2)
	assert.Equal(t, j.V("jarr").I(0).V("a").Int(), Int{1, true})
	assert.Equal(t, j.V("jarr").I(1).V("b").Int(), Int{2, true})

	assert.False(t, j.V("koko").Undefined())
	assert.True(t, j.V("no").Undefined())

	require.Len(t, j.V("sarr").Array().Values(), 2)
}

func TestBadArrays(t *testing.T) {
	j, err := NewJson(`{
		"a": null,
		"b": 123,
		"good": [1,2,3]
	}`)
	require.NoError(t, err)

	assert.Equal(t, 123, j.V("b").Int().Value)

	assert.False(t, j.V("a").Array().IsValid)
	assert.Len(t, j.V("a").Array().Values(), 0)

	assert.False(t, j.V("b").Array().IsValid)
	assert.Len(t, j.V("b").Array().Values(), 0)

	assert.True(t, j.V("a").I(0).Undefined())
}

func TestMarshal(t *testing.T) {
	j := Map{
		"key": "value",
	}

	str := j.Marshal(nil)
	assert.Equal(t, `{"key":"value"}`, str)
	str = j.MarshalPretty(nil)
	assert.Equal(t, `{
  "key": "value"
}`, str)

	j = Map{
		"bad": make(chan int),
	}

	str = j.Marshal(nil)
	assert.Equal(t, "", str)

	errCallbackCalled := false
	str = j.Marshal(func(err error) {
		errCallbackCalled = true
	})
	assert.Equal(t, "", str)
	assert.True(t, errCallbackCalled)
}

func TestNew(t *testing.T) {
	js, err := NewJson([]byte(`{"koko":"moko"}`))
	assert.NoError(t, err)
	assert.Equal(t, js.V("koko").String().Value, "moko")

	js, err = NewJson(`{"koko":"lala"}`)
	assert.NoError(t, err)
	assert.Equal(t, js.V("koko").String().Value, "lala")

	js, err = NewJson("{broken: }")
	assert.Error(t, err)
	assert.Equal(t, Json{}, js)

	js, err = NewJson("123")
	assert.NoError(t, err)
	assert.Equal(t, Json{float64(123), true}, js)

	js, err = NewJson(123)
	require.NoError(t, err)
	assert.Equal(t, Json{float64(123), true}, js)
}

func TestNewFromMap(t *testing.T) {
	jm1 := Map{
		"koko": "moko",
	}

	js1, err := NewJson(jm1)
	assert.NoError(t, err)
	assert.Equal(t, "moko", js1.V("koko").String().Value)

	jm1["koko"] = "lala"
	assert.Equal(t, "lala", jm1["koko"].(string))
	assert.Equal(t, "moko", js1.V("koko").String().Value)

	jm2 := Map{
		"koko": make(chan int),
	}
	js2, err := NewJson(jm2)
	assert.Error(t, err)
	assert.Equal(t, Json{}, js2)
}

func TestNewFromReader(t *testing.T) {
	reader := strings.NewReader(`{"koko": "moko"}`)

	j, err := NewJson(reader)
	assert.NoError(t, err)
	assert.Equal(t, "moko", j.V("koko").String().Value)
}

func TestUnmarshal(t *testing.T) {
	type Pixel struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	var pixel Pixel

	j, err := NewJson(Map{
		"pix": Map{
			"x": 123,
			"y": 456,
		},
	})
	require.NoError(t, err)

	err = j.V("pix").Unmarshal(&pixel)
	assert.NoError(t, err)
	assert.Equal(t, pixel, Pixel{123, 456})
}

func TestMarshalArray(t *testing.T) {
	jarr := Array{
		[]interface{}{
			1,
			"koko",
			nil,
		},
		true,
	}

	bytes, err := json.Marshal(jarr.Values())
	assert.NoError(t, err)
	assert.Equal(t, `[1,"koko",null]`, string(bytes))
}

func TestUnmarshalArray(t *testing.T) {
	str := `[{"x":10}, {"y":20}]`

	var jarr Json
	err := json.Unmarshal([]byte(str), &jarr)
	assert.NoError(t, err)

	require.False(t, jarr.Undefined())
	require.True(t, jarr.Array().IsValid)
	require.Len(t, jarr.Array().Values(), 2)
	assert.Equal(t, 10, jarr.Array().Values()[0].V("x").Int().Value)
	assert.Equal(t, Int{0, false}, jarr.Array().Values()[0].V("z").Int())

	notArrStr := `{"a": 1}`
	err = json.Unmarshal([]byte(notArrStr), &jarr)
	assert.NoError(t, err)
	assert.False(t, jarr.Array().IsValid)
}
