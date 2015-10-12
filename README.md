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

* Implements most the CTP API defined in the [CTP Data model and API
specification, v2.13](http://htmlpreview.github.io/?https://github.com/cloudsecurityalliance/ctpd/blob/master/client/CTP-Data-Model-And-API.html)
[(pdf)](https://github.com/cloudsecurityalliance/ctpd/blob/master/client/CTP-Data-Model-And-API.pdf)
defined by Cloud Security Alliance.

* Implements a ['back office' API](http://htmlpreview.github.io/?https://github.com/cloudsecurityalliance/ctpd/blob/master/client/CTP-Admin-API.html) 
that allows to update the database managed by ctpd. __This extra API is not part of the official CTP specification__.

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
For more details, see [the data model and API](http://htmlpreview.github.io/?https://github.com/cloudsecurityalliance/ctpd/blob/master/client/CTP-Data-Model-And-API.html)
[(pdf)](https://github.com/cloudsecurityalliance/ctpd/blob/master/client/CTP-Data-Model-And-API.pdf).

The actual implementation of the API is left to the choice of the service
provider, and the source code provided here is just an example of such an
implementation, without any normative value.

Licence
-------

ctpd is copyright 2015 Cloud Security Alliance EMEA (cloudsecurityalliance.org)
and is licensed under the Apache License, Version 2.0, as described
in the LICENCE file.

ctpd contains an optional demo client which uses:

* bootstrap, licenced under the MIT licence
as described in [here](https://github.com/twbs/bootstrap/blob/master/LICENSE).

* jstree, licenced under the MIT licence
as described in [here](https://github.com/vakata/jstree/#license--contributing).
