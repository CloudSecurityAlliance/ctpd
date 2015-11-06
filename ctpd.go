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

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
    "path"
	"github.com/cloudsecurityalliance/ctpd/server"
	"github.com/cloudsecurityalliance/ctpd/server/ctp"
)

const CTPD_VERSION = 0.1

var (
    configFileFlag string
    versionFlag bool
    logfileFlag string
    colorFlag bool
    debugVMFlag bool
    clientFlag string
    helpFlag bool
)

func init() {
	flag.StringVar(&configFileFlag, "config", "/path/to/file", "Specify an alternative configuration file to use.")
	flag.BoolVar(&versionFlag, "version", false, "Print version information.")
	flag.StringVar(&logfileFlag, "log-file", "", "Store logs in indicated file instead of standard output.")
	flag.BoolVar(&colorFlag, "color-logs", false, "Print logs with color on terminal.")
	flag.BoolVar(&debugVMFlag, "debug-vm", false, "Enable CTPScript virtual machine debugging output in logs.")
	flag.StringVar(&clientFlag, "client", "", "Set path to optional lightweight embedded javasciprt client. If empty, client is dissabled.")
	flag.BoolVar(&helpFlag, "help", false, "Print help.")
}

func main() {
	var ok bool
	var conf ctp.Configuration

	flag.Parse()

	if versionFlag {
		fmt.Println("ctpd version", CTPD_VERSION)
		fmt.Println(" Copyright 2015 Cloud Security Alliance EMEA (cloudsecurityalliance.org).")
		fmt.Println(" ctpd is licensed under the Apache License, Version 2.0.")
		fmt.Println(" see http://www.apache.org/licenses/LICENSE-2.0")
		fmt.Println("")
		return
	}

	if helpFlag {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
		return
	}

    if logfileFlag!="" {
        file, err := os.Create(logfileFlag)
        if err!=nil {
            log.Fatalf("Could not open %s, %s", logfileFlag, err.Error())
        }
        defer file.Close()
        log.SetOutput(file)
    }

	if configFileFlag == "/path/to/file" {
		conf, ok = ctp.SearchAndLoadConfigurationFile()
	} else {
		conf, ok = ctp.LoadConfigurationFromFile(configFileFlag)
	}

	if !ok {
        ctp.Log(nil,ctp.INFO,"No configuration file was loaded, using defaults.")
        conf = ctp.ConfigurationDefaults
	}

    if colorFlag {
        conf["color-logs"]="true"
    }

    if clientFlag!="" {
        conf["client"]=clientFlag
    }

    if debugVMFlag {
        conf["debug-vm"]="true"
    }

	if conf["client"] != "" {
		http.Handle("/", http.FileServer(http.Dir(conf["client"])))
	}

    if !ctp.IsMongoRunning(conf) {
        log.Fatal("Missing mongodb.")
    }

	http.Handle(conf["basepath"], server.NewCtpApiHandlerMux(conf))
	if conf["tls_use"] != "" && conf["tls_use"] != "no" {
		if conf["tls_use"] != "yes" {
			log.Fatal("Configuration: tls_use must be either 'yes' or 'no'")
		}
		if conf["tls_key_file"] == "" || conf["tls_cert_file"] == "" {
			log.Fatal("Missing tls_key_file or tls_cert_file in configuration.")
		}
		ctp.Log(nil,ctp.INFO,"Starting ctpd with TLS enabled at %s", conf["listen"])
		log.Fatal(http.ListenAndServeTLS(conf["listen"], conf["tls_cert_file"], conf["tls_key_file"], nil))
	} else {
		ctp.Log(nil,ctp.INFO,"Starting ctpd at %s", conf["listen"])
		log.Fatal(http.ListenAndServe(conf["listen"], nil))
	}
}
