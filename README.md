# jsn 

[![GoDoc](https://godoc.org/github.com/michael-go/go-jsn/jsn?status.svg)](https://godoc.org/github.com/michael-go/go-jsn/jsn)

![logo](logo.png)

### WORK IN PROGRESS - INTERFACES WILL CHANGE

## Features
* Flexibale constructor - consumes `string`, `[]byte`, `io.Reader` and any `interface{}` that is `json` Marshalable
* Safe (no `panic()`) access to keys of deeply nested JSONs (including arrays)
* value getters return a struct `{Value, IsValid}` instead of multiple return params for easier inlining (`Value` defaulting to a sensible default)
* Implementing `sql.Scanner` & `sql.Valuer` for easy integration with JSON columns
* Cute `jsn.Map` wrapper to `map[string]interface{}` for composition of arbiterary JSON objects 
* Easier iteration over JSON arrays
* Various other helper methods for happier life

## Usage

### Import

```go
import "github.com/michael-go/go-jsn/jsn"
```

### General

* `jsn.Json` represents any valid JSON including: map, array, bool, number, string, or null
* `jsn.Map` represents only a JSON map. Useful for composing JSON objets

### Accessing nested values:

any sub-element (map value or array element) of a `Json` is a `Json`.
* use `Get(key string)` or it's shortcut `K(key string)` to get a map sub-element by key
* use `I(index int)` to get an array sub-element by index
* calling the above methods on a non map/array element will just return an empty Json which is basically equivalent to Javascipt's `undefined`, but here you can safely call it's methods without `panic`-ing or `null`-dereferencing
* to get the actual value of a leaf element, use one of `Json`'s value methods `String()`, `Int()`, `Int64()`, `Float64()`, `Bool()` depending on the expected type. Each returns a struct with the typed `Value` and an `IsValid` field that will be false if the actual type is different or if the `Json` object itself is "undefined"

an example:
```go
j, err := jsn.NewJson(`{
    "go": {
        "pher": [10, "koko", {
            "lang": 20,
            "name": "gogo"
        }],
        "path": "/home"
    }
}`)
if err != nil {
    panic(err)
}

// dive through keys & array indexes:
fmt.Println(j.K("go").K("pher").I(2).K("name").String().Value)
// => "gogo"

// asking for the wrong type:
fmt.Println(j.K("go").K("path").Bool().Value)
// => false (deafult value for bool)
fmt.Printf("%#v\n", j.K("no").K("pher").Bool())
// => jsn.Bool{Value:false, IsValid:false}

// going into a non-existing path
fmt.Println(j.K("no").K("pher").Int().Value)
// => 0 (default value for int)
fmt.Printf("%#v\n", j.K("no").K("pher").Int())
// => jsn.Int{Value:0, IsValid:false}
```

Safely getting the "gogo" value in the first example above via vanilla `Go` would look like 🙈:
```go
var j map[string]interface{}
err := json.Unmarshal([]byte(`{
    "go": {
        "pher": [10, "koko", {
            "lang": 20,
            "name": "gogo"
        }],
        "path": "/home"
    }
}`), &j)
if err != nil {
    panic(err)
}

var value string
if g, ok := j["go"].(map[string]interface{}); ok {
    if p, ok := g["pher"].([]interface{}); ok && len(p) > 2 {
        if e, ok := p[2].(map[string]interface{}); ok {
            if n, ok := e["name"].(string); ok {
                value = n
            }
        }
    }
}
fmt.Println(value)
// => gogo
```

## Composing JSON objects
`jsn.Map` is just a fancy alias to `map[string]interface{}`, but sometimes the little things in life make all the difference. 
It also has some convinience methods for easirer marshling.

```go
type Pixel struct {
    X int
    Y int
}

jm := jsn.Map{
    "songs": []jsn.Map{
        {"hip": "hop"},
        {"hoo": "ray"},
    },
    "time":     time.Now(),
    "location": Pixel{X: 13, Y: 37},
}

fmt.Println(jm)
// => {"location":{"X":13,"Y":37},"songs":[{"hip":"hop"},{"hoo":"ray"}],"time":"2017-09-08T14:40:23.903861328+03:00"} 
fmt.Println(jm.Pretty())
// => same as above but pretty
```

**Note**: because `interface{}` can be anything, it is possible to create a `jsn.Map` that is not a valid JSON - i.e. `json.Marshal()` will fail on it. This can happen is a value is not marshalable - see https://golang.org/pkg/encoding/json/#Marshal.
In such case `String()` & `Pretty()` will return an empty string

---

## Inspired by these great projects:
* Also dealing with arbitrary JSON:
    * https://github.com/Jeffail/gabs
    * https://github.com/bitly/go-simplejson
* https://github.com/guregu/null

