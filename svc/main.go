// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"flag"
	"math/rand"
	"reflect"
	"strings"
	"svc/exit"
	"time"
)

type Endpoint struct {
	Values   map[string]interface{}
	Function func(string)
}

var endpoints map[string]*Endpoint

func init() {
	rand.Seed(time.Now().UnixNano())
	endpoints = map[string]*Endpoint{"lightning": &Endpoint{map[string]interface{}{}, lightning},
		"hello": &Endpoint{map[string]interface{}{"name": "", "count": 0}, hello}}
}

func main() {
	for _, endpoint := range endpoints {
		for key, value := range endpoint.Values {
			switch v := value.(type) {
			case int:
				endpoint.Values[key] = flag.Int(key, v, "")
			case string:
				endpoint.Values[key] = flag.String(key, v, "")
			case float64:
				endpoint.Values[key] = flag.Float64(key, v, "")
			case bool:
				endpoint.Values[key] = flag.Bool(key, v, "")
			default:
				endpoint.Values[key] = flag.String(key, exit.Sprint(v), "")
			}
		}
	}
	route := flag.String("endpoint", "", "")
	method := flag.String("method", "", "")
	flag.Parse()
	validateRoute(*route)
	route = &[]string{strings.ToLower(*route)}[0]
	method = &[]string{strings.ToUpper(*method)}[0]
	for _, endpoint := range endpoints {
		for key, _ := range endpoint.Values {
			endpoint.Values[key] = reflect.ValueOf(endpoint.Values[key]).Elem().Interface()
		}
	}

	//fmt.Println(endpoints[*route].Values)
	//fmt.Println(*route)
	//fmt.Println(*method)
	endpoints[*route].Function(*method)
}
func validateRoute(route string) {
	found := false
	for name, _ := range endpoints {
		if name == route {
			found = true
			break
		}
	}
	if !found {
		exit.Failure("Bad route")
	}
}
func hello(method string) {
	switch method {
	case "GET":
		exit.Success(`{"greeting":"Hello!","countSquared":-1}`)
	case "POST":
		values := endpoints["hello"].Values
		name := values["name"]
		count := values["count"]
		exit.Failure(`{"greeting":"Hello, %s!","countSquared":%d}`, name.(string), count.(int)*count.(int))
	default:
		exit.Failure("Bad method")
	}
}
func lightning(method string) {
	switch method {
	case "GET":
		strike := (rand.Intn(2) == 0)
		if !strike {
			exit.Success(`{"strike":false,"lat":0.0,"lon":0.0}`)
		}
		exit.Success(`{"strike":true,"lat":%.3f,:"lon":%.3f}`, (rand.Float32()*180.0)-90.0, (rand.Float32()*360.0)-180.0)
	default:
		exit.Failure("Bad method")
	}
}
func printJson(i interface{}) {
	dat, err := json.MarshalIndent(i, " ", "   ")
	println(string(dat), err)
}
