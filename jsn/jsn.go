// Package jsn was created to make arbitrary JSON handling more fun.
//
// more info at in the repo README: https://github.com/michael-go/go-jsn
package jsn

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// String carries a string .Value if .IsValid
type String struct {
	Value   string
	IsValid bool
}

// Int carries an int .Value if .IsValid
type Int struct {
	Value   int
	IsValid bool
}

// Int64 carries an int64 .Value if .IsValid
type Int64 struct {
	Value   int64
	IsValid bool
}

// Float64 carries a float64 .Value if .IsValid
type Float64 struct {
	Value   float64
	IsValid bool
}

// Bool carries a bool .Value if .IsValid
type Bool struct {
	Value   bool
	IsValid bool
}

// Array represents a JSON array, if .IsValid
// the actual data is opaque and can be accessed via Array's methods
type Array struct {
	elements []interface{}
	IsValid  bool
}

// Elements returns an array of Json items.
// Defaults to empty array if !(.IsValid)
func (a Array) Elements() []Json {
	if !a.IsValid || a.elements == nil {
		return []Json{}
	}

	values := make([]Json, len(a.elements))
	for i := 0; i < len(a.elements); i++ {
		values[i] = Json{a.elements[i], true}
	}

	return values
}

/////////////////

// Json represents any valid JSON: map, array, bool, number, string, or null.
// it's an opaque structure, and the data can be accessed via it's methods
type Json struct {
	data   interface{}
	exists bool
}

// NewJson constructs a new Json object from a wide variety of sources:
// - a JSON string from a string, []byte, io.Reader
// - any interface{} that is json.Marshal-able
func NewJson(src interface{}) (js Json, err error) {
	var data interface{}

	switch src.(type) {
	case []byte:
		err = json.Unmarshal(src.([]byte), &data)
	case string:
		err = json.Unmarshal([]byte(src.(string)), &data)
	case io.Reader:
		err = json.NewDecoder(src.(io.Reader)).Decode(&data)
	default:
		var bytes []byte
		bytes, err = json.Marshal(src)
		if err != nil {
			break
		}

		err = json.Unmarshal(bytes, &data)
	}

	if err == nil {
		js = Json{data, true}
	}

	return
}

func (j Json) asMap() (m map[string]interface{}, ok bool) {
	if !j.exists {
		return nil, false
	}

	switch j.data.(type) {
	case map[string]interface{}:
		return j.data.(map[string]interface{}), true
	default:
		return nil, false
	}
}

func (j Json) asArray() (a []interface{}, ok bool) {
	if !j.exists {
		return nil, false
	}

	a, ok = j.data.([]interface{})
	return
}

// Exists returns true if it's a Json map and the key exists.
// returns false otherwise
func (j Json) Exists(key string) bool {
	m, ok := j.asMap()

	if !ok {
		return false
	}

	_, exists := m[key]
	return exists
}

// Get returns the nested Json value under a key.
// returns an empty Json{} if key doesn't exists, or if this isn't a map
func (j Json) Get(key string) Json {
	m, ok := j.asMap()

	if !ok {
		return Json{}
	}

	v, exists := m[key]
	return Json{v, exists}
}

// K is a shortcut for Get()
func (j Json) K(key string) Json {
	return j.Get(key)
}

// I returns a array element by index.
// if index is out of bounds, or if this is not an array, it returns an undefined Json{}
func (j Json) I(index int) Json {
	a, ok := j.asArray()

	if !ok {
		return Json{}
	}

	if index < 0 || index > len(a)-1 {
		return Json{}
	}

	return Json{a[index], true}
}

// IterMap calls the callback for every kay-value pair in a JSON map,
// and returns the number of keys iterated.
// If it's not a map value, this method will do nothing.
// Caller can break the loop by returning false from the callback.
func (j Json) IterMap(f func(key string, value Json) bool) int {
	m, ok := j.asMap()
	if !ok {
		return 0
	}

	count := 0
	for k, v := range m {
		count++
		if !f(k, Json{v, true}) {
			break
		}
	}

	return count
}

// Undefined returns true if this Json is undefined.
// in example result of .Get(key) with a key that doesn't exist.
// like in JS, Null() != Undefined().
func (j Json) Undefined() bool {
	return !j.exists
}

// Null returns true if this object represents a JSON null.
// like in JS, Null() != Undefined().
func (j Json) Null() bool {
	return j.exists && j.data == nil
}

// NullOrUndefined returns (.Null() || .Undefined())
func (j Json) NullOrUndefined() bool {
	return j.data == nil
}

func (j Json) String() String {
	if !j.exists {
		return String{}
	}

	if s, ok := (j.data).(string); ok {
		return String{s, true}
	}
	return String{}
}

func (j Json) Int64() Int64 {
	if !j.exists {
		return Int64{}
	}

	switch j.data.(type) {
	case float64:
		return Int64{int64(j.data.(float64)), true}
	case json.Number:
		v, err := j.data.(json.Number).Int64()
		return Int64{v, err == nil}
	default:
		return Int64{}
	}
}

func (j Json) Int() Int {
	v := j.Int64()

	return Int{int(v.Value), v.IsValid}
}

func (j Json) Float64() Float64 {
	if !j.exists {
		return Float64{}
	}

	switch j.data.(type) {
	case float64:
		return Float64{j.data.(float64), true}
	case json.Number:
		v, err := j.data.(json.Number).Float64()
		return Float64{v, err == nil}
	default:
		return Float64{}
	}
}

func (j Json) Bool() Bool {
	if !j.exists {
		return Bool{}
	}

	switch j.data.(type) {
	case bool:
		return Bool{j.data.(bool), true}
	default:
		return Bool{}
	}
}

func (j Json) Array() Array {
	a, ok := j.asArray()

	return Array{a, ok}
}

func (j Json) Raw() interface{} {
	return j.data
}

func (j Json) Unmarshal(target interface{}) error {
	var data interface{}
	if j.exists {
		data = j.data
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, target)
}

func (j Json) Marshal() (string, error) {
	buf, err := json.Marshal(j)
	if err != nil {
		return "", err
	}
	return string(buf), err
}

func (j Json) MarshalIndent(prefix, indent string) (string, error) {
	buf, err := json.MarshalIndent(j, prefix, indent)
	if err != nil {
		return "", err
	}
	return string(buf), err
}

func (j Json) Pretty() string {
	buf, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return ""
	}

	return string(buf)
}

func (j Json) StringifyIndent(prefix, indent string) string {
	buf, err := json.MarshalIndent(j, prefix, indent)
	if err != nil {
		return ""
	}

	return string(buf)
}

func (j Json) Stringify() string {
	buf, err := json.Marshal(j)
	if err != nil {
		return ""
	}

	return string(buf)
}

// implementing json.Marshaler interface
func (j Json) MarshalJSON() ([]byte, error) {
	if j.exists {
		return json.Marshal(j.data)
	} else {
		return json.Marshal(Json{}.data)
	}
}

// implementing the json.Unmarshler interface
func (j *Json) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &j.data)

	j.exists = (err == nil)
	return err
}

// implementing the sql.Scanner interface
func (j *Json) Scan(src interface{}) error {
	switch src.(type) {
	case []byte:
		return j.UnmarshalJSON(src.([]byte))
	default:
		return fmt.Errorf("unsupprted src type")
	}
}

// implementing the sql/driver.Valuer interface
func (j Json) Value() (driver.Value, error) {
	bytes, err := j.MarshalJSON()
	return driver.Value(bytes), err
}

func (j Json) Reader() io.Reader {
	//TODO: is ignoring the error here is ok?
	bytes, _ := json.Marshal(j)

	return strings.NewReader(string(bytes))
}

////

// Map is just an alias to `map[string]interface{}` for easier construction
// of json objects
type Map map[string]interface{}

// TODO: how to handle error? ..
func (m Map) Json() Json {
	j, _ := NewJson(m)
	return j
}

func (m Map) Raw() map[string]interface{} {
	return map[string]interface{}(m)
}

func (m Map) Marshal() (string, error) {
	buf, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(buf), err
}

func (m Map) MarshalIndent(prefix, indent string) (string, error) {
	buf, err := json.MarshalIndent(m, prefix, indent)
	if err != nil {
		return "", err
	}
	return string(buf), err
}

func (m Map) Pretty() string {
	buf, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return ""
	}

	return string(buf)
}

func (m Map) Stringify() string {
	buf, err := json.Marshal(m)
	if err != nil {
		return ""
	}

	return string(buf)
}

func (m Map) StringifyIndent(prefix, indent string) string {
	buf, err := json.MarshalIndent(m, prefix, indent)
	if err != nil {
		return ""
	}

	return string(buf)
}
