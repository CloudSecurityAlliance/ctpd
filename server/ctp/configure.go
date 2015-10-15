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
	"bufio"
	"log"
	"os"
	"os/user"
	"path"
	"regexp"
	"strings"
)

type Configuration map[string]string

var ConfigurationDefaults = Configuration{
	"listen":      ":8080",
	"basepath":    "/api/1.0/",
	"databaseurl": "localhost",
}

var validEntry1 = regexp.MustCompile(`^([a-zA-Z0-9_]+)\s*=\s*([^ "\t\r\n]+)$`)
var validEntry2 = regexp.MustCompile(`^([a-zA-Z0-9_]+)\s*=\s*"([^"]*)"$`)

func LoadConfigurationFromFile(fname string) (Configuration, bool) {
	info, err := os.Stat(fname)
	if err != nil {
		return nil, false
	}
	if (info.Mode() & 077) != 0 {
		Log(nil, ERROR, "Permissions 0%o for %s are too open.", info.Mode()&os.ModePerm, fname)
		log.Fatalf("Configuration file should not be readable or writable by other users.")
	}

	file, err := os.Open(fname)
	if err != nil {
		return nil, false
	}
	defer file.Close()

	conf := make(map[string]string)
	linecount := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		linecount++
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[0:i]
		}
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			if r := validEntry1.FindStringSubmatch(line); r != nil {
				conf[r[1]] = r[2]
			} else if r := validEntry2.FindStringSubmatch(line); r != nil {
				conf[r[1]] = r[2]
			} else {
				log.Fatalf("Error on line %d in %s", linecount, fname)
			}
		}
	}

	// sanitize
	for k, v := range ConfigurationDefaults {
		if conf[k] == "" {
			conf[k] = v
		}
	}
        // NOTE: basepath is never = "" here so it's ok to do len-1, should be fixed however
	if conf["basepath"][len(conf["basepath"])-1] != '/' {
		conf["basepath"] += "/"
	}

	Log(nil, INFO, "Loaded configuration from %s.", fname)

	return conf, true
}

func SearchAndLoadConfigurationFile() (Configuration, bool) {
	cwd, err := os.Getwd()
	if err == nil {
		fname := path.Join(cwd, "ctpd.conf")
		if c, r := LoadConfigurationFromFile(fname); r {
			return c, true
		}
	}

	usr, err := user.Current()
	if err == nil {
		fname := path.Join(usr.HomeDir, ".ctpd.conf")
		if c, r := LoadConfigurationFromFile(fname); r {
			return c, true
		}
	}

	return LoadConfigurationFromFile("/etc/ctpd.conf")
}
