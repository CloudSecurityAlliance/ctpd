Cloud Trust Protocol Daemon Prototype
=====================================

This repository contains a prototype server implementing the Cloud Security
Alliance's [Cloud Trust Protocol](https://cloudsecurityalliance.org/group/cloudtrust-protocol/).

:warning: For the latest release of the __ctpd__ source code please always
refer to: https://github.com/cloudsecurityalliance/ctpd

This prototype called __ctpd__ is a unix-style server written in [Go](http://golang.org) with mongodb as a database backend. It has been tested
on Ubundtu/Debian Linux and Mac OS X. The code of __ctpd__ is still in 'beta'
stage and is mainly intended for testing and research purposes.


The __ctpd__ server provided in this repository:

* Implements most the CTP API defined in the _CTP Data model and API
specification, v2.13_ defined by Cloud Security Alliance.

* Implements a 'back office' API that allows to update the database managed
by ctpd. __This extra API is not part of the official CTP specification__.

Trying ctpd
-----------

To compile and run **ctpd**, you will need to install:

* go (http://golang.org),
* mongodb (https://www.mongodb.org/) and
* its go backend (https://labix.org/mgo).

See the INSTALL file for more details on installing these dependencies.

Next, run the mongodb script `build_db.js` in the `tools/` subdirectory of
this source code repository:

```bash
mongo build_db.js
```

Launch __ctpd__ for a test drive as follows:

```bash
go run ctpd.go
```

By default, __ctpd__ runs on port 8080, and you can test that is working
with a simple curl command:

```bash
curl -H "Authorization: Bearer 1234" http://localhost:8080/api/1.0/
```

The result should look something like this:
```json
{
  "self": "http://localhost:8080/api/1.0/",
  "name": "",
  "annotation": "ctpd prototype server",
  "version": "0",
  "provider": "",
  "serviceViews": "http://localhost:8080/api/1.0/serviceViews",
  "metrics": "http://localhost:8080/api/1.0/metrics"
}
```
Note that the value "1234" above is not an example of a secure token and
was created by `build_db.js` for demonstration purposes only.

Using ctpd
----------

_More information should be provided here soon._

About CTP
---------
The Cloud Trust Protocol (CTP) is designed to be a mechanism by which cloud
service customers can ask for and receive information related to the security
of the services they use in the cloud, promoting transparency and trust.

The Cloud Security Alliance has defined a "CTP data model and API", which
specifies how monitoring information should be presented to cloud customers.
For more details, see [missing link]().

The actual implementation of the API is left to the choice of the service
provider, and the source code provided here is just an example of such an
implementation, without any normative value.

Licence
-------

Copyright 2015 Cloud Security Alliance EMEA (cloudsecurityalliance.org)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.