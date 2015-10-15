//    Copyright 2015 Cloud Security Alliance EMEA (cloudsecurityalliance.org)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ctp

import (
	"fmt"
	"log"
	"sync"
)

var logMutex sync.Mutex

type LogLevel uint

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

var color_tags = [...]string{
	"\033[34mDEBUG\033[0m",
	"\033[90mINFO\033[0m",
	"\033[93mWARNING\033[0m",
	"\033[91mERROR\033[0m",
}

func Log(c *ApiContext, level LogLevel, format string, v ...interface{}) {
	var id string

    if c==nil {
        id = "[*] "
    } else {
        if c.ColorLogs {
            id = fmt.Sprintf("[\033[35m%d\033[0m] ", c.Id)
            id += color_tags[level] + " "
        } else {
            id = fmt.Sprintf("[%d] ", c.Id)
        }
    }

	logMutex.Lock()
	defer logMutex.Unlock()

	log.Printf(id+format, v...)
}
