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
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

var config Config

type Config struct {
	Cmd     string `json:"cmd"`
	Routing Routes `json:"routing"`
}

type Routes map[string]Methods
type Methods map[string]Params
type Params map[string]interface{}

func main() {
	dat, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal("Error reading config file: " + err.Error())
	}
	if err = json.Unmarshal(dat, &config); err != nil {
		log.Fatal("Error unmarshalling config file: " + err.Error())
	}
	printJson(config)

	routes := []RouteData{}
	for route, methods := range config.Routing {
		for method, _ := range methods {
			fmt.Println(route, method)
			routes = append(routes, RouteData{
				Verb:    method,
				Path:    "/" + route,
				Handler: handleRequest,
			})
		}
	}

	server := &Server{}
	if err = server.Configure(routes); err != nil {
		log.Fatal("Error configuring server: " + err.Error())
	}
	done, err := server.Start()
	if err != nil {
		log.Fatal("Error starting server: " + err.Error())
	}
	fmt.Println("Started server")
	fmt.Println(<-done)
}

func handleRequest(c *gin.Context) {
	method := c.GetString("_method")
	route := c.GetString("_route")
	body := map[string]interface{}{}
	if method == "PUT" || method == "POST" {
		if err := c.BindJSON(&body); err != nil {
			c.String(400, fmt.Sprintf(`{"error":"%s"}`, err.Error()))
			return
		}
	}
	confiParams := config.Routing[route][method]
	{
		for bodyK, bodyV := range body {
			found := false
			for confK, confV := range confiParams {
				if confK == bodyK {
					found = true
					if reflect.TypeOf(confV).String() != reflect.TypeOf(bodyV).String() {
						c.String(400, fmt.Sprintf(`{"error":"The value of [%s] should be type [%s] not [%s]"}`, confK, reflect.TypeOf(confV), reflect.TypeOf(bodyV)))
						return
					}
				}
			}
			if !found {
				c.String(400, fmt.Sprintf(`{"error":"Incorrect paramter [%s]"}`, bodyK))
				return
			}
		}
	}
	printJson(confiParams)
	flags := []string{"-endpoint=" + route, "-method=" + method}
	for k, v := range body {
		flags = append(flags, fmt.Sprintf("-%s=%#v", k, v))
	}
	fmt.Println(flags)
	dat, err := exec.Command(config.Cmd, flags...).Output()
	code := map[bool]int{true: 200, false: 400}[err == nil]
	res := strings.TrimSpace(string(dat))
	fmt.Println(res)
	c.String(code, res)
}

func printJson(i interface{}) {
	dat, err := json.MarshalIndent(i, " ", "   ")
	fmt.Println(string(dat), err)
}
